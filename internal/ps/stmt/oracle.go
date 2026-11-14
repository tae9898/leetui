package stmt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"leetui/internal/ps/lcfmt"
	"leetui/internal/ps/lcjudge"
	"leetui/internal/ps/problem"
	"leetui/internal/ps/runner"
)

// OracleResult는 FillOracle의 결과 — notice 문구의 재료.
type OracleResult struct {
	Targets int      // .out이 없어 대상이 된 .in 수
	Filled  []string // .out을 채운 케이스명 (3, ex1 …)
	Failed  []string // 실패한 케이스 — "3.in(사유)" 형태
	Cached  int      // 캐시 히트로 네트워크 없이 채운 수
}

// Summary는 notice 한 줄 요약.
func (r OracleResult) Summary() string {
	if r.Targets == 0 {
		return "오라클: 채울 .out 없음 (.out 없는 .in이 대상)"
	}
	s := fmt.Sprintf("오라클: %d개 중 %d개 채움", r.Targets, len(r.Filled))
	if len(r.Failed) > 0 {
		s += fmt.Sprintf(", %d개 실패 — %s", len(r.Failed), strings.Join(r.Failed[:min(3, len(r.Failed))], " "))
		if len(r.Failed) > 3 {
			s += fmt.Sprintf(" 외 %d개", len(r.Failed)-3)
		}
	}
	return s
}

// FillOracle은 .out이 없는 .in 전부를 LC 공식 정답으로 채운다 (D4 기본:
// 덮어쓰기 없음 — 이미 있는 .out은 손으로 쓴 정답으로 간주한다).
// 오라클이 채운 케이스명은 문제 디렉터리의 .oracle 파일에 남겨 파일 패널
// 표식(◆)으로 구분한다.
func FillOracle(ctx context.Context, root, slug string, p problem.Problem) (OracleResult, error) {
	j := &lcjudge.Judge{CachePath: oracleCachePath(root, p.Site, slug)}
	return fillOracle(ctx, root, slug, p, j)
}

func fillOracle(ctx context.Context, root, slug string, p problem.Problem, j *lcjudge.Judge) (OracleResult, error) {
	if p.Site != "lc" {
		return OracleResult{}, fmt.Errorf("오라클은 LeetCode만 지원")
	}
	res := OracleResult{}
	q, err := CachedQuestion(ctx, root, p.Site, slug)
	if err != nil {
		return res, err
	}
	meta, err := lcfmt.ParseMeta(q.MetaData)
	if err != nil {
		return res, err
	}
	if ok, reason := lcfmt.Supported(meta, q.Content); !ok {
		return res, fmt.Errorf("미지원: %s", reason)
	}
	cases, err := runner.Cases(p.Dir)
	if err != nil {
		return res, err
	}
	var targets []runner.Case
	for _, c := range cases {
		if c.HasOut || filepath.Base(c.In) == "custom.in" {
			continue // 손으로 쓴 .out 보존 / 커스텀 입력은 스크래치
		}
		targets = append(targets, c)
	}
	res.Targets = len(targets)
	if len(targets) == 0 {
		return res, nil
	}
	auth, err := lcjudge.LoadAuth(root)
	if err != nil {
		return res, err
	}

	inputs := make([][]string, len(targets))
	names := make([]string, len(targets))
	var valid []int // 로컬 변환을 통과한 케이스 인덱스
	for i, c := range targets {
		names[i] = strings.TrimSuffix(filepath.Base(c.In), ".in")
		data, err := os.ReadFile(c.In)
		if err != nil {
			res.Failed = append(res.Failed, fmt.Sprintf("%s.in(읽기 실패: %v)", names[i], err))
			continue
		}
		lits, err := lcfmt.FromStdin(meta, string(data))
		if err != nil {
			res.Failed = append(res.Failed, fmt.Sprintf("%s.in(%v)", names[i], err))
			continue
		}
		inputs[i] = lits
		valid = append(valid, i)
	}
	if len(valid) == 0 {
		return res, nil // 전부 로컬 변환 실패 — 네트워크 갈 필요 없음
	}
	validInputs := make([][]string, len(valid))
	for k, i := range valid {
		validInputs[k] = inputs[i]
	}

	results, err := j.Expected(ctx, auth, q, validInputs)
	if err != nil {
		return res, err
	}

	var marked []string
	for k, cr := range results {
		i := valid[k]
		if cr.Err != nil {
			res.Failed = append(res.Failed, fmt.Sprintf("%s.in(%s)", names[i], cr.Err))
			continue
		}
		out, err := lcfmt.ToStdout(meta, cr.Answer)
		if err != nil {
			res.Failed = append(res.Failed, fmt.Sprintf("%s.in(%v)", names[i], err))
			continue
		}
		if err := os.WriteFile(filepath.Join(p.Dir, names[i]+".out"), []byte(out), 0o644); err != nil {
			return res, err
		}
		res.Filled = append(res.Filled, names[i])
		if cr.Cached {
			res.Cached++
		}
		marked = append(marked, names[i])
	}
	if err := markOracle(p.Dir, marked); err != nil {
		return res, fmt.Errorf(".oracle 기록 실패: %w", err)
	}
	return res, nil
}

func oracleCachePath(root, site, slug string) string {
	return filepath.Join(root, ".stmt", site, slug+".oracle.json")
}

// markOracle은 오라클이 채운 케이스명을 .oracle 파일(한 줄 하나)에 합친다.
func markOracle(dir string, names []string) error {
	if len(names) == 0 {
		return nil
	}
	path := filepath.Join(dir, ".oracle")
	set := map[string]bool{}
	if data, err := os.ReadFile(path); err == nil {
		for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if l = strings.TrimSpace(l); l != "" {
				set[l] = true
			}
		}
	}
	for _, n := range names {
		set[n] = true
	}
	out := make([]string, 0, len(set))
	for n := range set {
		out = append(out, n)
	}
	sort.Strings(out)
	return os.WriteFile(path, []byte(strings.Join(out, "\n")+"\n"), 0o644)
}

// OracleMarked는 dir/.oracle에 기록된 케이스명 집합 — 파일 패널 표식용.
func OracleMarked(dir string) map[string]bool {
	set := map[string]bool{}
	data, err := os.ReadFile(filepath.Join(dir, ".oracle"))
	if err != nil {
		return set
	}
	for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			set[l] = true
		}
	}
	return set
}
