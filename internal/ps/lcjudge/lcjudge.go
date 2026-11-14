// Package lcjudge는 LeetCode interpret_solution 오라클이다. 스텁 코드로
// 내 입력(data_input)을 제출하면 채점기가 공식 풀이의 정답
// (expected_code_answer)을 돌려주는 비공식 동작을 이용한다.
//
// 비공식 엔드포인트라 LC가 바꾸면 깨진다 — 호출부(stmt.FillOracle)는
// 오류를 notice로만 전파해 나머지 기능에 영향이 없게 한다.
// 쿠키는 어디에도 로그로 남기지 않는다.
package lcjudge

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"leetui/internal/ps/lcref"
)

// Auth는 leetcode.com 세션 쿠키.
type Auth struct {
	Session string // LEETCODE_SESSION
	CSRF    string // csrftoken
}

// LoadAuth는 env(LC_SESSION/LC_CSRF)를 우선 쓰고, 없으면 $root/.lc-auth를 읽는다.
// 권한이 0600이 아니면 쿠키가 새는 것을 막기 위해 거부한다.
func LoadAuth(root string) (Auth, error) {
	if s, c := os.Getenv("LC_SESSION"), os.Getenv("LC_CSRF"); s != "" && c != "" {
		return Auth{Session: s, CSRF: c}, nil
	}
	path := filepath.Join(root, ".lc-auth")
	st, err := os.Stat(path)
	if err != nil {
		return Auth{}, fmt.Errorf("오라클 미설정 — %s 를 준비하거나 LC_SESSION/LC_CSRF 환경변수 (README 참고)", path)
	}
	if st.Mode().Perm() != 0o600 {
		return Auth{}, fmt.Errorf("%s 권한이 %s — chmod 600 후 다시", path, st.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Auth{}, err
	}
	var a Auth
	if err := json.Unmarshal(data, &a); err != nil || a.Session == "" || a.CSRF == "" {
		return Auth{}, fmt.Errorf("%s 형식 오류 — {\"session\":…,\"csrf\":…} JSON 필요", path)
	}
	return a, nil
}

// Judge는 interpret_solution 클라이언트. Base/HTTP/Backoff/PollBudget은
// httptest 주입용.
type Judge struct {
	Base       string        // ""면 https://leetcode.com
	HTTP       *http.Client  // nil면 http.DefaultClient
	CachePath  string        // .stmt/lc/<slug>.oracle.json — 같은 입력 재전송 방지
	Delay      time.Duration // 요청 사이 대기 (rate limit 대응). 0이면 400ms
	Backoff    time.Duration // 폴링 첫 대기. 0이면 1s, 최대 10s까지 지수 증가
	PollBudget time.Duration // 폴링 전체 예산. 0이면 60s
}

// CaseResult는 케이스 하나의 정답 조회 결과.
type CaseResult struct {
	Answer string // LC 정답 리터럴 (실패 시 "")
	Cached bool   // 캐시 히트 — 네트워크 없이 채움
	Err    error  // 이 케이스만 실패 — 입력 형식 오류 등
}

// ErrAuth는 401/403 — 쿠키 갱신이 필요하다.
var ErrAuth = errors.New("쿠키 만료 또는 인증 거부(403) — .lc-auth 갱신")

// ErrRate는 429 — 남은 요청도 실패할 것이므로 전체를 중단한다.
var ErrRate = errors.New("요청 과다(429) — 잠시 뒤 다시")

// Expected는 케이스별 LC 리터럴 입력(파라미터당 한 줄)에 대해 공식 정답
// 리터럴을 받아온다. 반환 슬라이스는 inputs와 같은 길이이고, 두 번째
// 반환값은 전체를 중단시키는 치명 오류(인증/과다요청/네트워크)다.
func (j *Judge) Expected(ctx context.Context, auth Auth, q lcref.Question, inputs [][]string) ([]CaseResult, error) {
	if strings.TrimSpace(q.CodeSnippet) == "" {
		return nil, fmt.Errorf("cpp 스텁 없음 — .stmt/lc/%s.json 캐시 삭제 후 재시도", q.Slug)
	}
	base := j.Base
	if base == "" {
		base = "https://leetcode.com"
	}
	hc := j.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	delay := j.Delay
	if delay == 0 {
		delay = 400 * time.Millisecond
	}

	cache := j.loadCache()
	results := make([]CaseResult, len(inputs))
	var need []int
	for i, in := range inputs {
		if in == nil {
			results[i] = CaseResult{Err: fmt.Errorf("로컬 변환 실패")}
			continue
		}
		key := cacheKey(in)
		if v, ok := cache[key]; ok {
			results[i] = CaseResult{Answer: v, Cached: true}
			continue
		}
		need = append(need, i)
	}
	for k, i := range need {
		if k > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		ans, err := j.one(ctx, base, hc, auth, q, inputs[i])
		if err != nil {
			if errors.Is(err, ErrAuth) || errors.Is(err, ErrRate) {
				return nil, err // 전체 중단 — 남은 요청도 실패할 것
			}
			results[i] = CaseResult{Err: err}
			continue
		}
		results[i] = CaseResult{Answer: ans}
		cache[cacheKey(inputs[i])] = ans
	}
	if err := j.saveCache(cache); err != nil {
		return results, fmt.Errorf("오라클 캐시 저장 실패: %w", err)
	}
	return results, nil
}

// one은 케이스 하나를 제출하고 폴링해 정답을 받는다.
func (j *Judge) one(ctx context.Context, base string, hc *http.Client, auth Auth, q lcref.Question, input []string) (string, error) {
	body, _ := json.Marshal(struct {
		Lang       string `json:"lang"`
		QuestionID string `json:"question_id"`
		TypedCode  string `json:"typed_code"`
		DataInput  string `json:"data_input"`
	}{"cpp", q.QuestionID, q.CodeSnippet, strings.Join(input, "\n")})
	url := base + "/problems/" + q.Slug + "/interpret_solution/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", base+"/problems/"+q.Slug+"/")
	req.Header.Set("Origin", base)
	req.Header.Set("x-csrftoken", auth.CSRF)
	req.Header.Set("Cookie", "csrftoken="+auth.CSRF+"; LEETCODE_SESSION="+auth.Session)

	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return "", ErrAuth
	case resp.StatusCode == http.StatusTooManyRequests:
		return "", ErrRate
	case resp.StatusCode/100 != 2:
		return "", fmt.Errorf("interpret_solution 응답 %s", resp.Status)
	}
	var out struct {
		InterpretID string `json:"interpret_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("interpret_id 파싱 실패: %w", err)
	}
	if out.InterpretID == "" {
		return "", fmt.Errorf("interpret_id 없음")
	}
	return j.poll(ctx, base, hc, out.InterpretID)
}

// poll은 /submissions/detail/<id>/check/ 를 지수 백오프로 돈다.
func (j *Judge) poll(ctx context.Context, base string, hc *http.Client, id string) (string, error) {
	backoff := j.Backoff
	if backoff == 0 {
		backoff = time.Second
	}
	budget := j.PollBudget
	if budget == 0 {
		budget = 60 * time.Second
	}
	deadline := time.Now().Add(budget)
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/submissions/detail/"+id+"/check/", nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Referer", base+"/")
		resp, err := hc.Do(req)
		if err != nil {
			return "", err
		}
		var out struct {
			State              string   `json:"state"`
			ExpectedCodeAnswer []string `json:"expected_code_answer"`
		}
		err = json.NewDecoder(resp.Body).Decode(&out)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("check 파싱 실패: %w", err)
		}
		switch out.State {
		case "SUCCESS":
			if len(out.ExpectedCodeAnswer) == 0 || out.ExpectedCodeAnswer[0] == "" {
				return "", fmt.Errorf("정답 없음 — 입력 형식 확인")
			}
			return out.ExpectedCodeAnswer[0], nil
		case "PENDING", "STARTED":
			// 아직 채점 중
		default:
			return "", fmt.Errorf("채점 상태 %q", out.State)
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("채점 시간 초과(%s)", budget)
		}
		wait := backoff
		if wait > 10*time.Second {
			wait = 10 * time.Second
		}
		backoff *= 2
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func cacheKey(input []string) string {
	sum := sha256.Sum256([]byte(strings.Join(input, "\n")))
	return hex.EncodeToString(sum[:])
}

func (j *Judge) loadCache() map[string]string {
	m := map[string]string{}
	if j.CachePath == "" {
		return m
	}
	data, err := os.ReadFile(j.CachePath)
	if err != nil {
		return m
	}
	_ = json.Unmarshal(data, &m)
	return m
}

func (j *Judge) saveCache(m map[string]string) error {
	if j.CachePath == "" || len(m) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(j.CachePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(j.CachePath, data, 0o644)
}
