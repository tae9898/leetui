# pstui

`~/ps` 연습 환경(plan.md 8절, fish `ps*` 함수)을 대체하는 TUI. 
`log.tsv` 포맷과 문제 디렉터리 규약은 기존 fish 환경과 100% 호환되어 둘을 번갈아 쓸 수 있다.
스펙: [CLI-ENV.md](CLI-ENV.md) · 계획: [TUI-PLAN.md](TUI-PLAN.md)

## 설치

```sh
make install   # ~/go/bin/pstui
```

fish에서 PATH 등록: `fish_add_path ~/go/bin`

## 실행

```sh
pstui              # TUI ($PS_ROOT 또는 ~/ps)
pstui --root ~/ps  # 루트 명시
```

의존: `g++`(빌드/테스트), `$EDITOR`(기본 nvim). fish/awk/diff는 쓰지 않는다.

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
| Workbench | `j/k` `e` `r` `F` | 파일 이동 / 편집 / 실행 / fast 빌드 |
| | `d` `t` `a` `o` `S` | 기록 / 재도전 / 케이스 추가 / .out 편집 / .start 재설정 |
| Run | `j/k` `h/l` `d` | 케이스 이동 / diff·stderr·컴파일 탭 / 결과 기록 |
| 플랜 피커 | `h/l` `u` `/` `⏎` `m` | 주차 / 미풀만 / 검색 / 시작 / 수동 폼 |
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
