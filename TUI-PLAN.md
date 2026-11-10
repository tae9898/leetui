# PS 연습 환경 TUI 재구현 계획

작성일: 2026-11-11 · 기준 문서: `CLI-ENV.md`(환경 규약), `plan.md` 8절(원본 스펙) · 상태: 계획 단계 (구현 미착수)

## 0. 확정 제약과 범위

| 항목 | 결정 |
|---|---|
| 스택 | Go + bubbletea (Charm 생태계: bubbles, lipgloss) |
| 구현 방식 | 네이티브 재구현. fish 함수를 호출하지 않고 모든 동작을 Go로 직접 구현 |
| 호환성 | `log.tsv` 포맷과 문제 디렉터리 규약은 `CLI-ENV.md`와 100% 동일. 기존 fish 함수 5종(psnew/psretry/psrun/psdone/psdue)과 같은 `~/ps`를 번갈아 써도 서로의 기록을 그대로 읽고 쓴다 |
| 외부 실행 의존 | `g++`, 편집기(`$EDITOR`, 기본 `nvim`)만. `timeout`, `diff`, `date`, `awk`, `fish`는 쓰지 않는다 |
| 비범위 | fish 함수 수정, nvim 키맵 수정, log.tsv 컬럼 추가, 문제 디렉터리 규약 변경, DB 도입 |

### 호환성 원칙 (구현 전체에 적용)

1. **쓰기는 fish와 바이트 단위로 동일하게.** TUI가 쓴 `log.tsv` 행과 `.start`는 fish가 쓴 것과 구별할 수 없어야 한다.
2. **읽기는 fish보다 관대하게.** fish가 만든 모든 산출물을 읽을 수 있어야 하며, 손상된 행(컬럼 부족 등)은 건너뛰되 TUI가 죽지 않는다.
3. **새 정보는 기존 파일 밖에 둔다.** 유형별 통계처럼 fish 환경에 없는 정보는 `log.tsv`에 컬럼을 추가하지 않고 `main.cpp` 헤더 주석을 파싱해 얻는다. TUI 전용 설정이 필요하면 `$PS_ROOT/.pstui/` 아래에 둔다 (현재 계획에서는 필요 없음).

## 1. 환경 조사 결과 (2026-11-11 기준)

| 항목 | 값 | 비고 |
|---|---|---|
| Go | 1.22.2 (`/usr/bin/go`), `GOTOOLCHAIN=auto` | bubbletea 최신(v1.3.10)은 Go 1.24 요구 → 아래 버전 고정 |
| g++ | 13.3.0 | `bits/stdc++.h`, `_GLIBCXX_DEBUG`, sanitizer 사용 가능 |
| fish | 3.7.0 | 상호 운용 테스트에 사용 (TUI 런타임 의존 아님) |
| nvim | 있음 | 기본 편집기 |
| `~/ps` | lc/ hr/ task/ mock/ template.cpp log.tsv(0바이트) | 테스트 디렉터리는 삭제된 상태 |

**의존성 고정 (Go 1.22 빌드 가능 조합, go.mod 직접 확인 완료)**

| 모듈 | 버전 | go 지시문 |
|---|---|---|
| github.com/charmbracelet/bubbletea | v1.3.4 | go 1.18 |
| github.com/charmbracelet/bubbles | v0.20.0 | go 1.18 |
| github.com/charmbracelet/lipgloss | v1.1.0 | go 1.18 |

bubbletea v1.3.5 이상과 bubbles v0.21.0은 go 1.23 이상을 요구한다. 대안은 `go.mod`에 `go 1.24`를 적고 `GOTOOLCHAIN=auto`로 툴체인을 내려받는 것이지만 네트워크 의존이 생기므로 기본안은 위 고정 조합이다. 그 외 서드파티 의존은 두지 않는다 (라인 diff는 직접 구현).

## 2. 아키텍처

### 2.1 레이어

```
┌──────────────────────────────────────────────┐
│  cmd/pstui            진입점, 플래그, 서브커맨드  │
├──────────────────────────────────────────────┤
│  internal/tui         bubbletea 모델/뷰/키맵     │  ← 도메인을 호출만 함
├──────────────────────────────────────────────┤
│  internal/ps  (도메인, TUI 비의존, 전부 단위테스트)│
│    config    PS_ROOT / EDITOR / 템플릿 경로       │
│    pslog     log.tsv 파싱·append·due·통계        │
│    problem   디렉터리 스캔, new, retry, .start, 헤더│
│    runner    컴파일, 케이스 실행, 타임아웃, 판정    │
│    linediff  기대/실제 출력 라인 diff            │
└──────────────────────────────────────────────┘
```

- `internal/ps/*`는 bubbletea를 import하지 않는다. 호환성 테스트가 이 레이어에 집중된다.
- `internal/tui`는 `internal/ps`의 함수를 `tea.Cmd`로 감싸 비동기 실행하고, 결과를 `tea.Msg`로 받는다.
- 서브커맨드 없이 실행하면 TUI, `pstui due` 같은 서브커맨드는 TUI 없이 도메인만 호출한다 (M5, 선택).

### 2.2 디렉터리 구성 (이 저장소 `/root/study/tui`)

```
tui/
├── go.mod                      module pstui, go 1.22
├── Makefile                    build / test / install / lint
├── cmd/pstui/main.go
├── internal/ps/
│   ├── config/config.go
│   ├── pslog/   log.go  due.go  stats.go  *_test.go  testdata/
│   ├── problem/ scan.go new.go retry.go start.go header.go *_test.go
│   ├── runner/  compile.go run.go cases.go *_test.go testdata/
│   └── linediff/diff.go *_test.go
├── internal/tui/
│   ├── app.go                  루트 모델: 화면 라우팅, 전역 키, 크기
│   ├── keys.go  styles.go  msgs.go
│   └── screen/  home.go problems.go workbench.go run.go done.go new.go retry.go log.go stats.go help.go
├── test/interop/               fish ↔ Go 상호 운용 테스트 (fish 없으면 skip)
├── CLI-ENV.md  plan.md  TUI-PLAN.md
```

### 2.3 도메인 계약 (호환성 핵심 — 구현 시 이 표를 그대로 테스트로 옮긴다)

#### config

| 항목 | 규칙 |
|---|---|
| `PS_ROOT` | 환경변수, 없으면 `~/ps`. `--root` 플래그가 최우선 |
| 템플릿 | `$PS_ROOT/template.cpp`. 없으면 psnew/psretry 비활성 + 홈 화면에 경고 |
| 편집기 | `$EDITOR`, 없으면 `nvim`. 문제 디렉터리를 cwd로 하여 `main.cpp` 인자로 실행 |
| 컴파일러 | `PATH`의 `g++`. `PS_CXX`로 바이너리 경로만 교체 가능, 플래그는 교체 불가 |

#### pslog — `log.tsv`

| 항목 | fish 동작 | Go 구현 |
|---|---|---|
| 위치 | `$PS_ROOT/log.tsv` | 동일 |
| 행 형식 | `printf '%s\t...\t%s\n'` 7컬럼, 헤더 없음 | 7컬럼을 탭으로 join 후 `\n`. 단일 `write` 호출 |
| 열기 | `>>` | `O_APPEND\|O_CREATE\|O_WRONLY`, 0644. 파일 끝이 `\n`이 아니면 먼저 `\n`을 보정 |
| 날짜 | `date +%F` (로컬 시간대) | `time.Now().Format("2006-01-02")` 로컬 |
| 재도전일 | ok→`-`, fail→`date -d "+3 days"`, again→`+7 days` | `AddDate(0,0,3)` / `AddDate(0,0,7)`. (DST 없는 KST에서는 fish와 동일) |
| 소요분 | `math -s0 "(now-start)/60"` | `(now-start)/60` 정수 나눗셈. 음수면 0 |
| 메모 | 인자 그대로 (탭 불가) | 탭·CR·LF를 공백으로 치환, 양끝 공백 trim. 비어 있어도 7번째 컬럼은 존재 |
| 파싱 | awk `-F'\t'` | 행별 `SplitN(line, "\t", 7)`. 컬럼이 6개 미만이면 건너뛰고 경고 카운트. 빈 행 무시. 7컬럼 초과(메모에 탭)는 나머지를 메모로 합침 |
| 인코딩 | UTF-8 | 동일. BOM 없음 |

#### pslog — due 조회 (psdue)

| 항목 | fish 동작 | Go 구현 |
|---|---|---|
| 키 | `$2 "/" $3` | `site + "/" + problem` |
| 우선 행 | 파일 순서상 마지막 행 | 동일 (map 덮어쓰기) |
| 필터 | `due != "-" && due <= today` (문자열 비교) | 동일하게 문자열 비교 (`YYYY-MM-DD`는 사전순 = 시간순) |
| 정렬 | `sort` (출력 행 전체 사전순) | `(due, key, result)` 바이트 순 |
| 출력 | `due key result` | 같은 세 값을 구조체로 반환, TUI는 표로, `pstui due`는 fish와 같은 문자열로 출력 |

#### problem — 디렉터리 규약

| 항목 | fish 동작 | Go 구현 |
|---|---|---|
| 경로 | `$PS_ROOT/<site>/<id>-<name>/` | 동일. 스캔 시 `<site>`는 PS_ROOT 바로 아래의 숨김 아닌 디렉터리 전부(lc/hr/task/mock에 한정하지 않음), 문제는 그 아래 디렉터리 |
| psnew | `mkdir -p`; main.cpp 없을 때만 템플릿 복사; `touch 1.in 1.out`; `.start` 기록; nvim | 동일. touch는 기존 파일을 자르지 않는다. 입력 검증: site/id/name에 `/`, 공백, 선행 `.` 불가, 모두 비어 있으면 안 됨 |
| `.start` | `date +%s >` → 에포크 초 + `\n` | `fmt.Fprintf("%d\n")`. 읽을 때 공백 trim 후 정수 파싱. 없거나 깨지면 "시작 시각 없음" 상태로 취급 (에러로 중단하지 않음) |
| psretry | `mv main.cpp main.MMDD.cpp`; 템플릿 복사; `.start` 갱신; nvim | 동일한 파일명. `main.MMDD.cpp`가 이미 있으면 덮어쓰기 전 확인 대화상자 (fish는 묻지 않고 덮어씀). main.cpp가 없으면 백업 생략 |
| 헤더 | `// 문제:` `// 유형:` `// 목표 복잡도:` 세 줄 | 앞 10줄에서 `// 유형:` 뒤 텍스트를 유형으로 파싱 (쉼표 구분 다중 유형 허용). 통계 전용, 쓰기에는 사용 안 함 |
| 산출물 | `sol`, `fast` 바이너리가 문제 디렉터리에 생김 | 같은 이름 사용 (fish와 서로 덮어써도 무방) |

#### runner — 빌드와 테스트 (psrun)

| 항목 | fish 동작 | Go 구현 |
|---|---|---|
| 컴파일 | `g++ -std=c++17 -O2 -g -Wall -Wextra -Wshadow -fsanitize=address,undefined -D_GLIBCXX_DEBUG main.cpp -o sol` | 플래그 문자열 상수로 고정, 변경 불가. stderr 전체를 캡처해 뷰에 표시. 실패 시 테스트 단계 생략 |
| 빠른 빌드 | `g++ -std=c++17 -O2 main.cpp -o fast` | 별도 키로 제공 (`F`). 시간 측정 표시 |
| 케이스 선택 | `*.in` (fish 글롭 자연 정렬) | `*.in` 중 `.out` 짝이 있는 것. 이름을 자연 정렬 (1,2,10). `.out` 없는 `.in`은 "기대출력 없음"으로 표시하고 실행은 하되 판정 제외 |
| 실행 | `timeout 2 ./sol < f` | `exec.CommandContext` + 2초 context, `Setpgid`로 프로세스 그룹 kill. stdout/stderr 분리 캡처, 출력 상한 1MiB |
| 판정 | `diff -q` 바이트 비교 (파이프라인 상태 = diff 상태) | 1차: 바이트 동일이면 PASS. **차이점**: 타임아웃은 `TLE`, 비정상 종료(sanitizer 포함)는 `RE`로 별도 표시하며 FAIL로 집계. fish는 부분 출력이 우연히 일치하면 PASS로 치지만 TUI는 더 엄격하다 (판정 결과는 어떤 파일에도 기록되지 않으므로 호환성에 영향 없음) |
| diff 출력 | `diff - exp` | `linediff`로 라인 단위 diff. 줄 끝 공백/개행 차이만 있는 경우 별도 힌트("공백만 다름") 표시 |
| 경과 | `(now - .start)/60`분 | 동일. `.start` 없으면 `-` 표시 |
| 종료 상태 | FAIL 있으면 1 | `Result{AllPass bool}`; 서브커맨드 모드에서는 exit 1 |
| 케이스 없음 | 리터럴 `*.in`으로 실패 | "테스트 케이스 없음" 안내 + `1.in/1.out` 생성 키 제공 |

#### linediff

의존성 없이 Myers 알고리즘(또는 LCS 기반) 라인 diff를 150줄 내로 구현한다. 입력은 기대/실제 `[]string`, 출력은 `(op, line)` 열. 최대 2,000줄까지만 diff하고 넘으면 "첫 불일치 줄 n"만 표시한다.

### 2.4 비동기와 프로세스 제어

- 컴파일, 케이스 실행, 디렉터리 스캔, 로그 파싱은 모두 `tea.Cmd` 안에서 실행한다. UI 고루틴에서 파일 I/O를 하지 않는다.
- 실행 중인 케이스는 `CaseStartedMsg` → `CaseDoneMsg`로 하나씩 들어와 목록이 실시간으로 갱신된다. `Esc`로 취소하면 context cancel → 프로세스 그룹 kill.
- 편집기는 `tea.ExecProcess`로 터미널을 넘겨 실행하고, 복귀 시 해당 문제의 파일 목록과 `.start`를 다시 읽는다.
- 터미널 크기 변경(`tea.WindowSizeMsg`)에 모든 화면이 반응한다. 최소 80×24, 그보다 작으면 안내 문구만 표시.

### 2.5 에러 처리 정책

- 도메인 함수는 `error`를 반환하고 TUI는 상태 표시줄에 1줄로 보여준다. 패닉 없이 복구 가능해야 한다.
- 치명 조건(PS_ROOT 없음, 템플릿 없음, g++ 없음)은 홈 화면 상단 경고 배너로 표시하고 관련 동작만 비활성화한다.
- `log.tsv` 손상 행은 건너뛰고 개수를 로그 뷰어 하단에 표시한다. 절대 자동으로 재작성하지 않는다 (fish 함수가 같은 파일을 쓰므로).

## 3. 화면과 기능 설계

### 3.1 화면 흐름

```
 [Home] ──p──▶ [Problems] ──Enter──▶ [Workbench] ──r──▶ [Run]
   │ n            │ n                   │ d ──▶ [Done form]
   │ l            │                     │ t ──▶ [Retry confirm]
   │ s            │                     │ e ──▶ (편집기 실행 후 복귀)
   ▼              ▼                     │ a ──▶ (N+1.in/.out 생성 후 편집기)
 [New form]   [Log] [Stats]             └ F ──▶ fast 빌드
 어디서나: ? 도움말, Esc 뒤로, q 종료(홈에서), Ctrl+C 즉시 종료
```

### 3.2 화면별 명세

**Home (대시보드)**
- 상단: PS_ROOT, 오늘 날짜, 경고 배너.
- 왼쪽 "오늘 재도전" 표: psdue 규칙 그대로 `재도전일 / 사이트/문제 / 마지막 결과`. 선택 후 `Enter`→Workbench, `t`→바로 Retry.
- 오른쪽 "최근 기록" 10행: 날짜 / 문제 / 결과 / 분 / 메모(잘림).
- 하단 "이번 주": 푼 수, ok/fail/again 수, 평균 분.
- 키: `n` 새 문제, `p` 문제 목록, `l` 로그, `s` 통계, `r` 새로고침.

**Problems (문제 목록)**
- 모든 `<site>/<problem>` 디렉터리를 스캔해 표로 표시: 사이트, 문제, 마지막 결과, 마지막 기록일, 재도전일, `.start` 기준 진행 중 표시, 헤더 유형.
- `/` 검색(부분 일치), `1`~`4` 또는 탭으로 사이트 필터, `d`로 due만 보기.
- 로그에는 있지만 디렉터리가 없는 문제는 회색으로 표시(삭제된 디렉터리).

**Workbench (문제 작업)**
- 헤더: 경로, 경과 시간(`.start` 기준, 1분마다 갱신), 헤더 주석 3줄.
- 파일 패널: `main.cpp`, `N.in/N.out` 쌍(크기, 빈 파일 표시), `main.MMDD.cpp` 백업 목록.
- 이력 패널: 이 문제의 log.tsv 행 전부.
- 마지막 실행 결과 요약(세션 내 유지).
- 키: `e` 편집, `r` 실행, `F` fast 빌드, `d` 결과 기록, `t` 재도전, `a` 케이스 추가(다음 번호의 `.in/.out` 생성 후 편집기로 `.in` 열기), `o` `.out` 편집, `S` `.start` 지금으로 재설정(없을 때만 활성, 확인 필요).

**Run (빌드와 테스트)**
- 단계 표시: 컴파일 중 → 케이스 n/m 실행 중 → 완료.
- 왼쪽: 케이스 목록 `PASS/FAIL/TLE/RE/NOEXP`, 실행 시간(ms).
- 오른쪽 뷰포트: 선택 케이스의 diff(기대 vs 실제), stderr(sanitizer 메시지), 컴파일 로그 탭 전환(`Tab`).
- 하단: 통과 수, 경과 분, `d`로 바로 Done form 이동(결과 기본값: 전부 PASS면 `ok`, 아니면 `fail`).
- `Esc` 실행 취소/뒤로.

**Done form (결과 기록)**
- 필드: 결과(`ok/fail/again` 라디오), 메모(textinput, 탭·개행 입력 불가), 소요분(자동 계산, `.start` 없으면 직접 입력 필수).
- 미리보기: 실제로 append될 TSV 행을 탭을 `⇥`로 치환해 보여준다. 재도전일도 미리 표시.
- `Enter` 저장 → 홈/Workbench의 캐시 갱신. 이미 오늘 같은 문제에 기록이 있으면 경고만 하고 허용(fish와 동일하게 누적).

**New form (새 문제)**
- 필드: 사이트(기존 디렉터리 목록에서 선택 + 직접 입력), id, name. 경로 미리보기 `$PS_ROOT/lc/0056-merge-intervals`.
- 이미 존재하는 디렉터리면 "기존 문제 열기"로 전환(main.cpp 덮어쓰지 않음, `.start`는 갱신 여부 확인).
- 생성 후 편집기 실행, 복귀하면 Workbench.

**Retry confirm**
- 백업 파일명(`main.MMDD.cpp`)을 보여주고 확인. 이미 있으면 덮어쓰기 경고.

**Log (로그 뷰어)**
- log.tsv 전체 표, 최신이 위. `/` 검색, `o/f/a`로 결과 필터, 손상 행 수 표시. 읽기 전용.

**Stats (통계, M4)**
- 유형별(헤더 `// 유형:`) 시도 수, fail 비율, 평균 분. 주차별 푼 수. 유형이 없는 문제는 "미분류".
- plan.md 6절 "일주일에 한 번 유형별로 묶어 본다"를 대체한다.

**Help overlay**
- 현재 화면의 키맵. `bubbles/help` 사용.

### 3.3 알려진 한계 4종의 해결 (CLI-ENV.md 7절)

| 한계 | TUI 처리 |
|---|---|
| 여러 단어 메모 일부만 기록 | 폼 입력이라 발생하지 않음. 탭·개행만 치환 |
| `.start` 없을 때 에러 | 경과 `-` 표시, Done form에서 분 직접 입력, `S`로 재설정 |
| `*.in` 없을 때 실패 | 안내 + 케이스 생성 키 |
| 유형 컬럼 없음 | 헤더 주석 파싱으로 통계. log.tsv는 7컬럼 유지 |

## 4. 마일스톤

| M | 이름 | 산출물 | 완료 기준 |
|---|---|---|---|
| M0 | 스캐폴드 | go.mod(의존성 고정), 디렉터리 골격, Makefile(`build/test/vet/install`), `pstui --version` | `go build ./...`와 `go vet ./...` 통과. 빈 TUI가 뜨고 `q`로 종료 |
| M1 | 도메인 코어 | `internal/ps/{config,pslog,problem,runner,linediff}` + 단위 테스트 + `test/interop` | 2.3절 계약 표의 모든 행이 테스트로 존재하고 통과. 상호 운용 테스트(아래 5.2) 통과 |
| M2 | TUI 뼈대 | app 라우팅, Home(due 목록·최근 기록), Problems, Help, 편집기 실행 | fish로 기록한 log.tsv가 Home에 psdue와 같은 순서로 보임. 문제 선택→편집기→복귀 동작 |
| M3 | 작업 루프 | Workbench, Run(스트리밍·취소·diff), Done form, New form, Retry | "새 문제 → 편집 → 실행 → 기록 → 다음날 due에 표시 → 재도전 → again 기록" 전체 루프를 TUI만으로 완료. 그 결과 log.tsv를 fish `psdue`가 그대로 읽음 |
| M4 | 마무리 | Log 뷰어, Stats, 한계 4종 해결, 크기 변경·에러 배너, README, `make install` | 5.3절 완료 기준 전부 충족 |
| M5 (선택) | 헤드리스 서브커맨드 | `pstui due/run/done/new` | 출력·종료 코드가 fish 함수와 동일(`due` 문자열 동일, `run` exit 1). nvim 키맵에서 fish 대신 호출 가능 |

M0~M4 순서는 고정이고 M1은 TUI 없이 끝낸다. M5는 요청이 있을 때만 진행한다.

## 5. 테스트와 완료 기준

### 5.1 단위 테스트 (M1)

- `pslog`: 고정 fixture(`testdata/log.tsv`)로 파싱, 손상 행 건너뛰기, 메모의 탭 치환, append 바이트 비교(골든 파일), due 계산(ok/fail/again, 경계일 `<=`), 마지막 행 우선, 정렬.
- `problem`: 임시 디렉터리에서 new(기존 main.cpp 보존, touch 비파괴), retry(백업 이름, 템플릿 복사), `.start` 읽기/쓰기(개행 포함), 헤더 파싱, 스캔(숨김 제외, 자연 정렬).
- `runner`: `testdata/`의 작은 C++ 프로그램 4종(정답, 오답, 무한루프, 범위 초과)으로 PASS/FAIL/TLE/RE 판정. g++ 없으면 `t.Skip`.
- `linediff`: 동일/삽입/삭제/치환/공백만 차이.
- `tui`: 각 화면 모델의 `Update`에 키 메시지를 넣어 상태 전이 검증 (렌더링 스냅샷은 하지 않음).

### 5.2 상호 운용 테스트 (`test/interop`, fish 있을 때만 실행)

임시 `PS_ROOT`에서 fish 함수(`~/.config/fish/conf.d/ps.fish`를 source)와 Go 도메인을 교차 실행한다.

1. Go `New` → fish `psrun`(빈 main 템플릿 + 빈 1.in/1.out은 PASS) → fish `psdone fail 메모` → Go `Due()`가 그 행을 3일 뒤 due로 읽는다.
2. fish `psnew` → Go `Append(ok)` → fish `psdue` 출력이 Go `Due()`를 문자열로 포맷한 것과 바이트 동일.
3. fish `psdone`과 Go `Append`가 같은 입력으로 만든 행이 탭 위치·컬럼 수·개행까지 동일 (날짜는 같은 날에 실행).
4. Go `Retry` 후 fish `psrun`이 정상 동작하고 백업 파일명이 `main.MMDD.cpp`.
5. fish가 `.start`를 쓴 디렉터리에서 Go가 경과 분을 fish와 같은 값으로 계산.

fish 함수는 `nvim`을 실행하므로 테스트에서는 `function nvim; end`로 무력화한 뒤 source한다.

### 5.3 최종 완료 기준 (M4 종료 시 전부 충족)

- [ ] `go build`, `go vet`, `go test ./...` 통과. 서드파티 의존은 Charm 3종뿐.
- [ ] 바이너리가 `fish`, `timeout`, `diff`, `date`, `awk`를 호출하지 않는다 (`strings`/코드 검색으로 확인).
- [ ] 5.2 상호 운용 테스트 5종 통과.
- [ ] 컴파일 플래그 문자열이 `CLI-ENV.md` 4절과 문자 단위로 동일하다 (테스트로 고정).
- [ ] TUI만으로 "psnew→psrun→psdone→psdue→psretry→psdone again" 루프를 완료할 수 있다.
- [ ] 같은 `~/ps`에서 fish 함수와 TUI를 번갈아 써도 양쪽 모두 오류 없이 모든 기록을 보여준다 (수동 점검 1회, 기록 남김).
- [ ] CLI-ENV.md 7절의 한계 4종이 3.3절대로 처리된다.
- [ ] 80×24 터미널에서 모든 화면이 깨지지 않는다.
- [ ] 실행 취소 시 자식 프로세스가 남지 않는다 (`ps` 확인).
- [ ] 코드 내 TODO/stub/skip 없음. README에 설치(`make install` → `~/go/bin/pstui`)와 키맵 수록.

## 6. 리스크와 결정

| 리스크 | 대응 |
|---|---|
| bubbletea 최신판이 Go 1.24 요구 | 1절의 버전 고정. 추후 Go 업그레이드 시 `go.mod`만 올림 |
| fish `date -d "+3 days"`와 `AddDate`의 DST 차이 | KST에는 DST가 없어 동일. 테스트는 날짜 계산 함수에 고정 `now`를 주입 |
| 편집기 실행 중 터미널 상태 | `tea.ExecProcess` 사용. 복귀 후 전체 재렌더 |
| sanitizer 바이너리가 느려 2초 초과 | fish와 같은 2초 유지(규약). TLE로 표시하고 `F` fast 빌드로 재확인하도록 안내 |
| 같은 날 두 번째 retry의 백업 충돌 | 확인 대화상자(덮어쓰기/취소). 파일명 규약은 유지 |
| 출력이 매우 큰 케이스 | 1MiB 상한, 초과 시 잘린 표시. 판정은 전체 바이트로 수행 |

## 7. 다음 단계

이 문서가 승인되면 M0부터 시작한다. 구현 착수 전 확인할 것 하나: 바이너리 이름 `pstui`와 설치 경로 `~/go/bin` (fish `fish_add_path`로 PATH 추가) 으로 진행해도 되는지.
