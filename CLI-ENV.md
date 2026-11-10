# PS 연습 CLI 환경 — 구현 내역 정리

작성일: 2026-11-11 · 원본 스펙: `plan.md` 8절 · 검증 상태: 전 함수 end-to-end 통과

## 1. 개요

42dot 코테 준비용 연습 환경. 모든 문제를 stdin/stdout 프로그램으로 풀고,
.in/.out 파일쌍으로 테스트하며, 실패한 문제를 3일 뒤 재도전하는 사이클을
fish 셸 함수 5종으로 운영한다. TUI(`TUI-PLAN.md` 참조)는 이 환경과
**완전 호환**되어야 하며 둘을 병행 사용할 수 있어야 한다.

## 2. 디렉터리 구조

```
~/ps/                        # $PS_ROOT (기본값 ~/ps)
├── template.cpp             # 새 문제 시작점
├── log.tsv                  # 활동 로그 (아래 포맷 참조)
├── lc/  hr/  task/  mock/   # 사이트/용도별 문제 디렉터리
└── <site>/<id>-<name>/      # 문제 1개 = 디렉터리 1개
    ├── main.cpp             # 풀이
    ├── 1.in  1.out          # 예제 입력/기대출력
    ├── 2.in  2.out          # 직접 만든 엣지케이스 (N은 1부터 연속)
    ├── .start               # 시작 epoch 초 (psnew/psretry가 기록)
    └── main.MMDD.cpp        # 재도전 시 백업 (psretry가 생성)
```

## 3. log.tsv 포맷 (호환성 핵심)

- 탭 구분 7컬럼, 헤더 없음, 문제당 여러 행 누적
- UTF-8, 마지막 컬럼(메모)은 빈 문자열 허용, 탭 포함 불가

| 컬럼 | 예시 | 설명 |
|---|---|---|
| 1 날짜 | `2026-11-11` | 기록일 `YYYY-MM-DD` |
| 2 사이트 | `lc` | 문제 디렉터리의 부모 이름 (lc/hr/task/mock) |
| 3 문제 | `0056-merge-intervals` | 문제 디렉터리 이름 |
| 4 결과 | `ok` | `ok` \| `fail` \| `again` |
| 5 소요분 | `25` | `(now - .start) / 60` 정수 |
| 6 재도전일 | `2026-11-14` | `YYYY-MM-DD` 또는 `-` |
| 7 메모 | `단조 스택 떠올리지 못함` | 자유 텍스트 |

### 재도전 규칙
- `ok` → 재도전일 `-` (목록에서 제외)
- `fail` → 기록일 +3일
- `again`(재도전 통과) → 기록일 +7일 최종 확인

### 조회 규칙 (psdue)
- 키 = `사이트/문제`, **같은 키의 마지막 행이 우선**
- 마지막 행의 재도전일이 `-`가 아니고 오늘 이하면 표시
- 재도전일 오름차순 정렬

## 4. fish 함수 스펙 (`~/.config/fish/conf.d/ps.fish`)

### psnew — 새 문제 시작
`psnew lc 0056 merge-intervals`
1. `$PS_ROOT/lc/0056-merge-intervals/` 생성 (이미 main.cpp 있으면 덮어쓰지 않음)
2. `template.cpp` → `main.cpp` 복사
3. `1.in`, `1.out` 빈 파일 touch
4. epoch 초를 `.start`에 기록
5. 해당 디렉터리로 이동 후 `nvim main.cpp`

### psretry — 재도전 (문제 디렉터리 안에서)
1. `main.cpp` → `main.MMDD.cpp` 백업
2. 템플릿으로 리셋, `.start` 갱신, nvim 실행

### psrun — 빌드와 테스트 (문제 디렉터리 안에서)
1. 컴파일 (아래 플래그 고정), 실패 시 즉시 종료
2. 모든 `*.in`에 대해 `timeout 2 ./sol < f` 실행, 대응 `.out`과 diff
3. 케이스별 `PASS n.in` / `FAIL n.in` + 실패 시 diff 출력
4. 경과 분 출력 (`(now - .start)/60`)
5. 하나라도 FAIL이면 exit 1

컴파일 플래그 (변경 금지 — 코테 함정 탐지용):
```
g++ -std=c++17 -O2 -g -Wall -Wextra -Wshadow \
    -fsanitize=address,undefined -D_GLIBCXX_DEBUG main.cpp -o sol
```
시간 초과 확인용 빠른 빌드는 별도: `g++ -std=c++17 -O2 main.cpp -o fast`

### psdone — 결과 기록
`psdone ok|fail|again [메모]` — 위 log.tsv 규칙대로 1행 append.
유효하지 않은 인자는 usage 출력 후 exit 1.

### psdue — 오늘 재도전할 문제
위 조회 규칙대로 `재도전일 사이트/문제 결과` 목록 출력.

## 5. 기타 파일

- `~/ps/template.cpp`: 헤더 주석 3줄(`문제/유형/목표 복잡도`) + `bits/stdc++.h` + fast IO main
- `~/.config/nvim/after/plugin/ps.lua`: `<space>r` → 저장 후 분할 터미널에서 psrun

## 6. 검증 이력 (2026-11-11)

| 항목 | 결과 |
|---|---|
| psnew 디렉터리/템플릿/.start 생성 | 통과 |
| psrun PASS/FAIL 판정, diff 출력, 경과, exit code | 통과 |
| psdone fail/again → due +3일/+7일 기록 | 통과 |
| psdue 날짜 필터 + 마지막 기록 우선 | 통과 |
| psretry 백업(main.1111.cpp) + 템플릿 리셋 | 통과 |
| nvim 키맵 로드 | 통과 |

테스트에 쓴 문제 디렉터리와 로그는 삭제 완료 (1주차 1번 문제 스포일러 방지).

## 7. 알려진 한계 (TUI에서 개선 여지)

- 메모를 따옴표 없이 여러 단어로 넘기면 일부만 기록됨 (printf 인자 문제)
- `.start` 없는 디렉터리에서 psrun/psdone 실행 시 경과 계산이 에러 메시지를 냄
- `*.in`이 하나도 없으면 루프가 리터럴 `*.in`을 대상으로 실패
- 로그에 유형(해시/DP 등) 컬럼이 없어 유형별 통계 불가 — 유형은 main.cpp 헤더 주석(`// 유형:`)에만 존재
