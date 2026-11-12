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
- 피커 `a` / 작업 화면 `A`: 지문의 예제 Input/Output을 숨김 스펙 파일 `.exN.in`/`.exN.out`으로 저장한다.
  실행에서는 제외되고 참고용(아직 시작하지 않은 문제면 디렉터리까지 만든다).

## HackerRank 방식으로 풀기

본시험 플랫폼이 HackerRank로 확정됐으므로(plan.md 6절 규칙 6) **모든 문제를 stdin 파싱 / stdout 출력으로** 푼다.
LeetCode 문제도 입력 형식을 직접 정의하고 main에서 파싱한다.

- `1.in`/`2.in`…: 자기 포맷의 실행 케이스 (작업 화면 `a`로 생성)
- `.exN.in/.out`: LC 예제 스펙 — 실행 안 됨, `.in` 만들 때 참고
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
| | `d` `t` `a` `A` `i` `o` `S` | 기록 / 재도전 / 케이스 추가 / 예제스펙 / 커스텀입력 / .out 편집 / .start 재설정 |
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
