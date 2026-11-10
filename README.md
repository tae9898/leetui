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

## 키맵

| 화면 | 키 | 동작 |
|---|---|---|
| Home | `j/k` | 재도전 목록 이동 |
| | `enter` | Workbench 열기 |
| | `t` | 재도전(백업+템플릿 리셋) |
| | `n` `p` `l` `s` | 새 문제 / 문제 목록 / 로그 / 통계 |
| | `r` `q` | 새로고침 / 종료 |
| Problems | `/` `tab` `1~4` `d` | 검색 / 사이트 필터 / due만 |
| Workbench | `e` `r` `F` | 편집 / 실행 / fast 빌드 |
| | `d` `t` `a` `o` `S` | 기록 / 재도전 / 케이스 추가 / .out 편집 / .start 재설정 |
| Run | `Tab` `d` `esc` | diff/stderr/컴파일 탭 / 결과 기록 / 취소·뒤로 |
| 어디서나 | `?` `ctrl+c` | 도움말 / 종료 |

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
