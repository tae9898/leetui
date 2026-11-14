# leetui

`~/ps` 연습 환경(plan.md 8절, fish `ps*` 함수)을 대체하는 TUI. 
`log.tsv` 포맷과 문제 디렉터리 규약은 기존 fish 환경과 100% 호환되어 둘을 번갈아 쓸 수 있다.
스펙: [CLI-ENV.md](CLI-ENV.md) · 계획: [TUI-PLAN.md](TUI-PLAN.md)

## 설치

```sh
make install   # ~/go/bin/leetui
```

fish에서 PATH 등록: `fish_add_path ~/go/bin`

## 실행

```sh
leetui              # TUI ($PS_ROOT 또는 ~/ps)
leetui --root ~/ps  # 루트 명시
```

의존: `g++`(빌드/테스트), `$EDITOR`(기본 nvim). fish/awk/diff는 쓰지 않는다.

## 지문과 예제 — 터미널 안에서

- 플랜 피커/작업 화면에서 `v`: 지문을 **마크다운**으로 변환해 `$EDITOR`(nvim)로 띄운다. LeetCode는
  지문(제목·난이도·본문)을 GraphQL로 가져와 `PS_ROOT/.stmt/<site>/<slug>.md`로 캐시한다 — 코드펜스·
  강조·목록이 살아 있고 nvim이 `filetype=markdown`으로 하이라이팅한다. 과제형 3종은 plan.md 7절 요구사항이 내장.
  nvim이므로 검색·스크롤·yank·분할 자유롭게. 캐시를 지우고 다시 받으려면 그 파일을 삭제.
- LC 문제를 만들면(플랜 피커 `⏎`, 수동 폼) **예제가 곧바로 실행 케이스가 된다** — 예제 입출력을
  leetui 표준 포맷(`ex1.in`/`ex1.out`…)으로 변환해 저장한다. 열자마자 `r`을 누르면 예제가 PASS/FAIL로 판정된다.
  지문 상단에는 "입력 형식 (leetui)" 섹션이 자동 삽입된다(HR의 Input Format 칸 대응).
- 피커 `a` / 작업 화면 `A`: 수동 재실행. 이미 있는 `exN.*`는 사용자가 손본 것으로 간주해 덮어쓰지 않고,
  원본 리터럴은 숨김 스펙 `.exN.in/.out`으로 계속 보존된다.

### leetui 표준 포맷

파라미터 순서대로 줄로 직렬화한다(`integer[]` = 크기 줄 + 값 줄, `character[][]` = `r c` + 행 문자열,
`TreeNode` = 노드 수 + 레벨순서 `null` 표기, void+제자리 수정은 수정된 파라미터를 출력 …).
정확한 규칙은 지문 상단 섹션 참고. **미지원 유형**(아래)은 `.exN` 참고 스펙만 남고 notice로 알린다:

- design 문제(`LRUCache`처럼 클래스 인터페이스)
- `double` 반환 (LC는 1e-5 오차 허용, leetui는 바이트 비교)
- 답 자체가 모음(`integer[][]`, `string[]`)이고 지문에 "any order"가 있는 문제 (정답이 여러 개)

## 오라클 — LC 공식 정답으로 .out 채우기 (작업 화면 `O`)

엣지케이스·랜덤 입력을 직접 만들고 `.out`이 없는 `.in` 전부를 LeetCode 채점기의 공식 정답으로 채운다.
히든 케이스에 가장 가까운 근사치. 오라클이 채운 `.out`은 파일 패널에 `◆` 표식이 붙는다(`.oracle` 파일에 기록).

- 대상: `.out` 없는 `.in` 전부 (손으로 쓴 `.out`은 절대 덮어쓰지 않는다. `custom.in`은 제외)
- 캐시: 같은 입력은 `.stmt/lc/<slug>.oracle.json`에 저장돼 다시 보내지 않는다
- **개인 연습용 소량 호출만** — 대량 자동 루프는 만들지 않는다 (비공식 엔드포인트·약관 회색지대)

### `.lc-auth` 설정 (오라클용 세션 쿠키)

1. 브라우저에서 leetcode.com 로그인 → 개발자도구(F12) → Application/Storage → Cookies
2. `LEETCODE_SESSION`과 `csrftoken` 값 복사
3. `~/ps/.lc-auth` 작성 (**권한 0600 필수** — 느슨하면 거부된다):

```sh
cat > ~/ps/.lc-auth <<'EOF'
{"session":"LEETCODE_SESSION 값","csrf":"csrftoken 값"}
EOF
chmod 600 ~/ps/.lc-auth
```

환경변수 `LC_SESSION`/`LC_CSRF`가 있으면 파일보다 우선한다. 쿠키는 로그·notice·에러에 절대 남지 않는다.
쿠키가 만료되면 `O` 실행 시 "쿠키 만료 — .lc-auth 갱신" notice가 뜬다(403).

## HackerRank 방식으로 풀기

본시험 플랫폼이 HackerRank로 확정됐으므로(plan.md 6절 규칙 6) **모든 문제를 stdin 파싱 / stdout 출력으로** 푼다.
LC 문제는 leetui 표준 포맷(위)을 쓴다 — 입력 형식을 지문이 정해주는 것이 실전과 같다.

- `ex1.in`…: 예제에서 자동 생성(`A`) · `1.in`/`2.in`…: 직접 만드는 케이스 (작업 화면 `a`)
- `.exN.in/.out`: LC 예제 원본 리터럴 스펙 — 실행 안 됨, 참고용
- 작업 화면 `i`: **커스텀 입력 실행** (HR "Run Custom Input" 대응) — 에디터로 `custom.in`을 쓰고 나오면
  그 파일만 fast 빌드로 돌려 stdout을 보여준다. `.out`이 없으므로 판정(NOEXP)에서는 제외된다.

## 문제 시작하는 두 가지 방법

1. **플랜 피커** (권장) — 홈에서 `n`. plan.md 5절의 전체 목록(1~4주차 + 과제형 + 여유, 66항)이
   주차·유형·상태(미풀/진행 중/✓완료/재도전)와 함께 뜬다. `⏎`만 누르면 디렉터리 생성 → 편집기 → 작업 화면까지.
   `h/l` 주차 필터, `u` 미풀만, `/` 검색.
2. **수동 폼** — 피커에서 `m`. 첫 필드에 URL(`https://leetcode.com/problems/merge-intervals/`)이나
   `56. Merge Intervals`, 슬러그를 붙여넣으면 나머지가 자동으로 채워진다. LeetCode URL은 번호까지
   GraphQL로 조회한다.

## 키맵

| 화면 | 키 | 동작 |
|---|---|---|
| Home | `j/k` | 재도전 목록 이동 |
| | `enter` | Workbench 열기 |
| | `t` | 재도전(백업+템플릿 리셋) |
| | `n` `p` `l` `s` | 플랜 피커 / 문제 목록 / 로그 / 통계 |
| | `r` `q` | 새로고침 / 종료 |
| Problems | `j/k` `h/l` `/` `d` | 이동 / 사이트 필터 / 검색 / due만 |
| Workbench | `j/k` `e` `v` `r` `F` | 파일 이동 / 편집 / 지문(nvim) / 실행 / fast 빌드 |
| | `d` `t` `a` `A` `O` `i` `o` `S` | 기록 / 재도전 / 케이스 추가 / 예제케이스 / 오라클(.out 채움) / 커스텀입력 / .out 편집 / .start 재설정 |
| Run | `j/k` `h/l` `d` | 케이스 이동 / diff·stderr·컴파일 탭 / 결과 기록 |
| 플랜 피커 | `h/l` `u` `/` `v` `a` `⏎` `m` | 주차 / 미풀만 / 검색 / 지문(nvim) / 예제채움 / 시작 / 수동 폼 |
| 새 문제 폼 | `enter` `tab` | 필드 진행 → 마지막에 생성 / 필드 이동 |
| 기록 폼 | `h/l` `j/k` | 결과 선택 / 필드 이동 |
| Log | `j/k` `h/l` `/` | 이동 / 결과 필터 / 검색 |
| 어디서나 | `?` `ctrl+c` | 도움말 / 종료 |

이동은 vim motion(`j/k/h/l`) 기준. 화살표 키도 동일하게 동작한다.

## 워크플로 (plan.md 6절 재도전 규칙 내장)

실패(`fail`)는 3일 뒤, 재도전 통과(`again`)는 7일 뒤 Home "오늘 재도전"에 자동 표시된다.
`ok`는 목록에서 빠진다. 판정은 fish보다 엄격하다: 타임아웃 `TLE`, 비정상 종료(sanitizer 포함) `RE`로 구분한다.

## 개발

```sh
make build test vet
```

- `internal/ps/*` — 도메인(외부 의존 없음), 상호 운용 테스트 `test/interop`(fish 있을 때만 실행)
- `internal/tui` — bubbletea 화면
- 의존성 고정: bubbletea v1.3.4 / bubbles v0.20.0 / lipgloss v1.1.0 (Go 1.22 호환)
