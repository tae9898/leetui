42dot 코딩테스트 준비 플랜
2026년 11월 10일 
1. 전제와 확인할 것
이 플랜은 "HackerRank에서 LeetCode Medium 수준 2~3문제를 C++로 푼다"는 가정 위에 서 있고, 그 가정은 아직 확인되지 않았습니다. 응시 안내 메일을 받는 즉시 아래 항목을 확인하고 3~4주차 계획을 맞추세요.
공고에서 확인된 것은 전형 순서뿐입니다: 서류전형 → 코딩테스트 → 화상면접(1시간 내외) → 대면 혹은 화상면접(3시간 내외). 직무별로 다르게 운영될 수 있다고 명시되어 있습니다 (AI Software Engineer 공고, MCU System Engineer 공고).
확인 항목
현재 가정
가정이 틀리면 바꿀 것
플랫폼
HackerRank (확정 — 2026-11-12 확인)
해당 없음. 응시 전 샘플 테스트 1회 필수
문항 수와 제한 시간
2~3문제, 90분
모의고사 세트 구성을 실제 값으로 변경
사용 언어
C++17 가능
C만 허용이면 6절 과제형 비중을 늘리고 STL 의존 풀이 축소
문제 형식
알고리즘 (stdin/stdout 또는 함수 완성)
과제형이면 7절을 2주차부터 병행
객관식 포함 여부
없음
C/OS 지식 문항 복습 추가
면접 라이브 코딩
있을 수 있음 (미확인)
없어도 "말하며 풀기"는 유지
응시 기한
약 4주 뒤
기간이 짧으면 3절의 압축안 적용
HackerRank로 확정되면 플랫폼이 제공하는 샘플 테스트를 응시 전에 반드시 한 번 돌려보세요. 에디터, 커스텀 입력 실행, 제출 방식에 익숙해지는 데 20분이면 됩니다.
2. 현재 수준 진단
메워야 할 구멍은 두 개입니다: STL을 손에 붙이는 도구 숙달과, 문제를 보고 유형을 떠올리는 패턴 훈련. 원래 플랜은 앞의 것만 문제로 봤지만, 알고리즘 풀이를 오래 쉬었다면 뒤의 것이 더 오래 걸립니다.
구분
내용
플랜에서의 처리
강점
C 문법, 포인터, 메모리
따로 연습하지 않음. 디버깅 속도의 기반
강점
파서, 상태머신, 구현형 문제
구현/파싱 유형을 확실한 득점원으로 만듦 (5절 보강 유형)
강점
pthread 동시성, 프로파일링
코테보다 면접용. 7절 과제형에서 활용
약점
STL과 modern C++ 활용 속도
1주차 전반에 집중, 이후 매 문제로 반복
약점
알고리즘 유형별 훈련 부족
2~3주차 유형별 Medium + 재도전 사이클
약점
시간 압박 하의 정확도
4주차 세트 모의고사 3회 이상
전략은 세 줄로 요약됩니다.
1. 모든 문제를 처음부터 stdin/stdout 프로그램으로 풀어, 입출력 처리를 별도 과제로 남기지 않는다.
2. 새 문제를 많이 푸는 것보다 못 푼 문제를 3일 뒤 빈 파일에서 다시 푸는 것을 우선한다.
3. 구현/파싱/비트 조작처럼 실무와 겹치는 유형은 "반드시 맞히는 문제"로 관리한다.
3. 4주 일정
평일 하루 2시간, 주말 하루 3~4시간 기준으로 주당 약 17시간, 4주 합계 약 68시간입니다. 매일의 첫 30분은 재도전 문제에 쓰고, 남은 시간에 새 문제를 풉니다.
주차
목표
새 문제
그 주에 끝나 있어야 하는 것
1주차
STL 숙달, 입출력 패턴, 연습 환경 구축
워밍업 Easy 12 + 연결리스트/트리 Easy 4
CLI 환경 완성, STL 치트시트를 안 보고 Easy를 15분 안에 풀이
2주차
유형별 Medium 전반
해시, 투포인터/슬라이딩, 배열, 구간, 스택, 이진탐색 (13문제)
유형별 템플릿 코드를 외워서 작성 가능
3주차
유형별 Medium 후반 + 보강 유형
BFS/DFS, 그래프, 백트래킹, DP, 비트, 구현/파싱 (16문제) + 과제형 3종
grid BFS와 인접리스트 BFS를 5분 안에 뼈대 작성
4주차
실전 적응
세트 모의고사 3~4회, 웹 에디터 풀이 8문제 이상
90분 세트에서 2문제 완주, 실패 문제 전부 재도전 완료
하루 루틴 (평일 2시간)
1. 0~30분: psdue로 나온 재도전 문제 1개를 빈 파일에서 풉니다.
2. 30~100분: 새 문제 2개. 문제당 30분 제한, 초과하면 풀이를 읽고 닫은 뒤 직접 구현합니다.
3. 100~120분: 그날 푼 문제의 유형, 막힌 지점, 실수 한 줄씩을 로그에 남깁니다.
1주차 상세
일차
할 일
1일
CLI 환경 구축 (8절), template.cpp 작성, 입출력 패턴 5종 직접 타이핑
2일
vector, sort와 비교자, pair, string 메서드. Easy 3문제
3일
unordered_map/set, map/set, lower_bound. Easy 3문제
4일
stack, queue, deque, priority_queue. Easy 3문제
5일
남은 Easy 3문제 + 재도전
주말
연결리스트/트리 Easy 4문제, 1주차 문제 중 15분 넘긴 것 전부 다시 풀기
기간이 2주로 줄어들 경우
1주차를 3일로 압축하고(Easy 8개만), Medium은 유형당 1문제씩 12개만 풉니다. 백트래킹과 DP는 각 1문제로 줄이고, 마지막 3일은 세트 모의고사 2회와 재도전에만 씁니다.
4. C++ 코테 도구
아래 세 블록을 1주차에 직접 타이핑해서 컴파일해 보고, 2주차부터는 보지 않고 씁니다. 기준은 C++17입니다.
입출력 패턴
#include <bits/stdc++.h>
using namespace std;

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    // 1) 개수 + 배열
    int n; cin >> n;
    vector<int> a(n);
    for (auto& x : a) cin >> x;

    // 2) cin >> 뒤에 줄 단위로 읽을 때: 남은 개행 제거
    cin.ignore(numeric_limits<streamsize>::max(), '\n');
    string line; getline(cin, line);

    // 3) 한 줄을 공백으로 분리
    stringstream ss(line);
    string tok;
    while (ss >> tok) { /* ... */ }

    // 4) 구분자로 분리 (예: "1,2,3")
    stringstream cs(line);
    while (getline(cs, tok, ',')) { /* stoi(tok) */ }

    // 5) 개수가 주어지지 않을 때: EOF까지
    int x;
    while (cin >> x) { /* ... */ }

    // 출력: endl 대신 '\n'
    cout << a[0] << '\n';
}
STL 치트시트
// vector
vector<int> v(n, 0);                       // 크기 n, 0으로 초기화
vector<vector<int>> g(r, vector<int>(c));  // r x c 2차원
v.push_back(x); v.pop_back(); v.back(); v.empty(); (int)v.size();

// 정렬과 탐색
sort(v.begin(), v.end());
sort(v.begin(), v.end(), greater<int>());
sort(v.begin(), v.end(), [](const auto& a, const auto& b) { return a.second < b.second; });
auto it = lower_bound(v.begin(), v.end(), x);   // x 이상인 첫 위치 (정렬된 상태에서)
int idx = it - v.begin();
v.erase(unique(v.begin(), v.end()), v.end());   // 정렬 후 중복 제거

// pair, tuple
pair<int,int> p = {1, 2};  p.first; p.second;
auto [a, b] = p;                                // 구조적 바인딩 (C++17)
tuple<int,int,int> t = {1, 2, 3};  auto [x, y, z] = t;

// 해시
unordered_map<string,int> m;  m[key]++;
if (m.count(key)) { }                           // 존재 확인은 count 또는 find
for (auto& [k, val] : m) { }
unordered_set<int> s;  s.insert(x); s.count(x); s.erase(x);

// 정렬된 컨테이너 (O(log n))
map<int,int> om;  set<int> os;
auto lo = os.lower_bound(x);                    // 멤버 함수 버전을 쓸 것
*os.begin(); *os.rbegin();                      // 최소, 최대

// 스택, 큐, 덱
stack<int> st;  st.push(x); st.top(); st.pop();
queue<int> q;   q.push(x); q.front(); q.pop();
deque<int> dq;  dq.push_front(x); dq.push_back(x); dq.pop_front(); dq.pop_back();

// 힙
priority_queue<int> maxH;                                   // 기본은 최대 힙
priority_queue<int, vector<int>, greater<int>> minH;        // 최소 힙
priority_queue<pair<int,int>, vector<pair<int,int>>, greater<>> pq;  // (거리, 노드) 최소 힙

// string
s.substr(pos, len); s.find("ab"); to_string(42); stoi("42"); stoll("9999999999");
isdigit(c); isalpha(c); tolower(c); c - '0'; c - 'a';
reverse(s.begin(), s.end());

// 기타
accumulate(v.begin(), v.end(), 0LL);            // 초기값 타입이 결과 타입
*max_element(v.begin(), v.end());
INT_MAX; LLONG_MAX; 1e9 + 7;
자주 밟는 함정
함정
증상
대처
비교자에 >= 또는 <= 사용
정렬 중 크래시 또는 오답 (정의되지 않은 동작)
항상 < 또는 >만 사용
m[key]로 존재 확인
조회만 해도 원소가 삽입되어 크기와 순회 결과가 바뀜
m.count(key) 또는 m.find(key)
v.size() - 1
빈 벡터에서 부호 없는 언더플로
(int)v.size() - 1
int 곱셈과 누적합
숨김 케이스에서만 오답
값 범위가 10^9 근처면 long long
accumulate(..., 0)
합이 int로 잘림
초기값을 0LL로
s.find 실패 값
-1과 비교하면 경고 또는 오동작
string::npos와 비교
stoi 범위 초과
예외로 종료
stoll 사용 또는 직접 파싱
cin >> 뒤 getline
빈 줄이 읽힘
cin.ignore로 개행 제거
순회 중 컨테이너 수정
iterator 무효화로 크래시
it = c.erase(it) 패턴 또는 별도 수집 후 삭제
unordered_map의 키로 pair
컴파일 에러 (해시 없음)
map 사용 또는 r * C + c로 정수 인코딩
endl 반복 출력
출력이 많을 때 시간 초과
'\n' 사용
재귀 DFS 깊이
큰 입력에서 스택 오버플로
깊이가 10^5 이상이면 명시적 스택 또는 BFS
5. 문제 목록
필수는 45문제(Easy 19, Medium 26)이고, 번호는 LeetCode 문제 번호입니다. 표의 순서대로 풉니다.
1주차: 워밍업 Easy 16문제
#
문제
익히는 것
1
Two Sum
unordered_map
217
Contains Duplicate
unordered_set
242
Valid Anagram
배열 카운팅
125
Valid Palindrome
투포인터, 문자 판별
704
Binary Search
이진탐색 경계 처리
121
Best Time to Buy and Sell Stock
한 번 스캔
20
Valid Parentheses
stack
88
Merge Sorted Array
뒤에서부터 병합
283
Move Zeroes
in-place 조작
136
Single Number
XOR
118
Pascal's Triangle
2차원 vector
53
Maximum Subarray
카데인 알고리즘
206
Reverse Linked List
포인터 조작 (강점 확인용)
21
Merge Two Sorted Lists
더미 노드
141
Linked List Cycle
빠른/느린 포인터
104
Maximum Depth of Binary Tree
재귀 DFS
2주차: Medium 전반 13문제
유형
#
문제
핵심
해시
49
Group Anagrams
정렬한 문자열을 키로
해시
347
Top K Frequent Elements
카운팅 + 힙
해시
560
Subarray Sum Equals K
누적합 + 해시
투포인터
3
Longest Substring Without Repeating Characters
슬라이딩 윈도우
투포인터
11
Container With Most Water
양끝에서 좁히기
투포인터
15
3Sum
정렬 + 투포인터, 중복 건너뛰기
배열
238
Product of Array Except Self
앞뒤 누적곱
배열
55
Jump Game
그리디
구간
56
Merge Intervals
정렬 후 병합
스택
739
Daily Temperatures
단조 스택
스택
150
Evaluate Reverse Polish Notation
스택 계산
이진탐색
153
Find Minimum in Rotated Sorted Array
조건 이진탐색
이진탐색
875
Koko Eating Bananas
파라메트릭 서치
3주차: Medium 후반과 보강 유형 16문제
유형
#
문제
핵심
BFS/DFS
200
Number of Islands
grid 탐색 뼈대
BFS/DFS
994
Rotting Oranges
다중 시작점 BFS, 레벨 단위 처리
BFS/DFS
102
Binary Tree Level Order Traversal
레벨 단위 BFS
그래프
207
Course Schedule
인접리스트, 위상정렬
백트래킹
78
Subsets
선택/비선택
백트래킹
46
Permutations
방문 배열
DP
198
House Robber
1차원 점화식
DP
322
Coin Change
최솟값 DP, 불가능 처리
비트
191
Number of 1 Bits
비트 순회 (Easy)
비트
190
Reverse Bits
시프트와 마스크 (Easy)
비트
338
Counting Bits
비트 + DP (Easy)
구현
8
String to Integer (atoi)
파싱, 오버플로 처리
구현
394
Decode String
스택 기반 파서
구현
54
Spiral Matrix
경계 관리 시뮬레이션
설계
146
LRU Cache
list + unordered_map, 면접 설명용으로도 필수
힙
973
K Closest Points to Origin
크기 K 힙
여유가 있으면
33 Search in Rotated Sorted Array, 300 Longest Increasing Subsequence, 62 Unique Paths, 79 Word Search, 743 Network Delay Time(다익스트라) 순으로 추가합니다.
4주차: 웹 에디터에서 풀 HackerRank 문제
Interview Preparation Kit에서 고릅니다. 목적은 영어 지문, 입출력 형식, 에디터 적응입니다.
난이도
문제
적응용
Sales by Match, Counting Valleys, Jumping on the Clouds, Repeated String
배열
2D Array - DS, Arrays: Left Rotation, New Year Chaos, Minimum Swaps 2
문자열/스택
Sherlock and Anagrams, Balanced Brackets
그래프/DP
BFS: Shortest Reach in a Graph, DFS: Connected Cell in a Grid, Max Array Sum
한국어 해설이 필요한 유형은 바킹독 "실전 알고리즘" 강의의 해당 회차만 골라 봅니다. 전체 수강은 이 일정에 넣지 않습니다.
6. 운영 규칙
규칙은 여섯 개이고, 이 중 재도전 규칙이 실력 향상에 가장 크게 기여합니다.
1. 30분 룰. Medium은 30분, Easy는 15분 안에 못 풀면 풀이를 읽습니다. 풀이를 닫은 뒤 직접 구현해서 통과시키고, 결과는 실패로 기록합니다.
2. 3일 뒤 재도전. 실패한 문제는 3일 뒤 빈 파일에서 다시 풉니다. 또 실패하면 다시 3일 뒤, 통과하면 7일 뒤 한 번 더 확인하고 목록에서 뺍니다.
3. 코드보다 복잡도 먼저. 제약 조건을 읽고 목표 복잡도를 주석 첫 줄에 적은 뒤 구현을 시작합니다. 아래 표를 기준으로 삼습니다.
4. 제출 전 엣지케이스 4종. 빈 입력, 원소 1개, 최대 크기, 전부 같은 값을 .in 파일로 만들어 돌립니다.
5. 말하며 풀기. 접근, 자료구조를 고른 이유, 복잡도를 소리 내어 설명하면서 풉니다. 면접에 라이브 코딩이 없더라도 풀이 설명 질문에 그대로 쓰입니다.
6. HackerRank 방식으로 풀기(2026-11-12 확정). 본시험 플랫폼이 HackerRank로 확정됐으므로 모든 문제를 stdin 파싱 / stdout 출력 방식으로 풉니다. LeetCode 문제도 입력 형식을 직접 정의하고 main에서 파싱합니다 — LC 예제 자동 채움(.exN.in)은 스펙 참고용이고, 실행 케이스(1.in, 2.in…)는 자기 포맷으로 직접 씁니다.
제약 조건에서 복잡도 역산
n의 최대값
허용되는 복잡도
떠올릴 접근
20 이하
O(2^n), O(n!)
백트래킹, 비트마스크 완전탐색
500 이하
O(n^3)
3중 루프, 플로이드
5,000 이하
O(n^2)
2중 루프, 2차원 DP
10^5 ~ 10^6
O(n log n), O(n)
정렬, 힙, 이진탐색, 해시, 투포인터
10^9 이상
O(log n), O(1)
값에 대한 이진탐색, 수식
풀이 순서 (문제당)
1. 지문과 제약 조건을 읽고 예제를 손으로 따라갑니다 (3분).
2. 유형과 목표 복잡도를 주석으로 적습니다 (2분).
3. 구현합니다 (15분).
4. 예제와 엣지케이스를 돌리고, 값 범위에서 int 오버플로를 확인합니다 (5분).
5. LeetCode 문제는 함수만 붙여 넣어 숨김 케이스를 통과시킵니다.
로그에 남길 것
날짜, 문제, 유형, 결과(ok/fail), 소요 시간, 막힌 지점 한 줄. 일주일에 한 번 로그를 유형별로 묶어 보고, 실패율이 높은 유형에서 다음 주 추가 문제를 고릅니다.
7. 과제형·임베디드 대비
임베디드 포지션은 알고리즘 대신 구현 과제가 나올 수 있으므로, 3주차 주말에 아래 세 가지를 각각 40분 제한으로 테스트 코드까지 작성합니다. C와 C++ 두 버전으로 한 번씩 짜 두면 언어 제한에도 대비됩니다.
과제
요구사항
확인할 포인트
링버퍼
고정 크기, push/pop, 가득 찼을 때 정책(거부 또는 덮어쓰기) 선택
full과 empty 구분, 인덱스 wrap, 단일 생산자-단일 소비자에서의 스레드 안전성 설명
고정 블록 메모리 풀
정적 배열 위에서 alloc/free, 동적 할당 없음
free list 구성, 이중 해제와 범위 밖 포인터 검출, 정렬(alignment)
프로토콜 파서
헤더, 길이, 페이로드, 체크섬 구조의 바이트 스트림 파싱
상태머신 구성, 조각난 입력 처리, 길이 필드 검증, 엔디안
각 과제는 assert 기반 테스트를 같은 파일에 작성하고, sanitizer를 켠 상태로 통과시킵니다. 채점자가 보는 것은 동작보다 경계 조건 처리와 인터페이스 설계인 경우가 많습니다.
면접 대비 지식 점검
코딩테스트 범위는 아니지만 객관식이 섞이거나 면접에서 풀이와 함께 물을 수 있는 항목입니다. 각 항목을 2분 안에 말로 설명할 수 있는지 4주차에 한 번 점검합니다.
영역
항목
C 언어
volatile, static의 세 가지 의미, const 포인터 조합, 구조체 패딩과 정렬, 비트필드, memcpy와 memmove 차이
정의되지 않은 동작
부호 있는 정수 오버플로, strict aliasing, 초기화되지 않은 변수, 범위 밖 시프트
시스템
엔디안, 스택과 힙, 인터럽트 컨텍스트 제약, 메모리 맵 I/O
동시성
mutex, semaphore, spinlock의 차이, 교착 조건 4가지, 우선순위 역전, 조건변수의 spurious wakeup
Modern C++
RAII, unique_ptr와 shared_ptr, 이동 의미론, 가상 함수 비용, std::array와 std::vector 선택 기준
8. CLI 연습 환경
원칙은 하나입니다: 모든 문제를 표준 입력을 읽어 표준 출력에 쓰는 프로그램으로 풀고, 테스트는 .in/.out 파일 쌍으로 관리합니다. LeetCode 문제도 Solution 함수와 직접 작성한 main()으로 로컬에서 통과시킨 뒤 함수만 사이트에 제출합니다.
디렉터리 구조
~/ps/
├── template.cpp           # 새 문제의 시작점
├── log.tsv                # 날짜 / 사이트 / 문제 / 결과 / 분 / 재도전일 / 메모
├── lc/                    # LeetCode
│   └── 0056-merge-intervals/
│       ├── main.cpp
│       ├── 1.in  1.out    # 예제
│       ├── 2.in  2.out    # 직접 만든 엣지케이스
│       └── .start         # 시작 시각 (psnew가 기록)
├── hr/                    # HackerRank
├── task/                  # 과제형 (링버퍼, 메모리 풀, 파서)
└── mock/                  # 모의고사 세트 (날짜별)
template.cpp
// 문제:
// 유형:
// 목표 복잡도:
#include <bits/stdc++.h>
using namespace std;

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    return 0;
}
fish 함수
~/.config/fish/conf.d/ps.fish에 저장합니다. GNU date와 timeout 기준이며, macOS에서는 coreutils의 gdate, gtimeout으로 바꿉니다.
set -q PS_ROOT; or set -gx PS_ROOT ~/ps

# psnew lc 0056 merge-intervals
function psnew -a site id name -d "새 문제 시작"
    set dir $PS_ROOT/$site/$id-$name
    mkdir -p $dir
    test -e $dir/main.cpp; or cp $PS_ROOT/template.cpp $dir/main.cpp
    touch $dir/1.in $dir/1.out
    date +%s > $dir/.start
    cd $dir
    nvim main.cpp
end

# 문제 디렉터리 안에서 실행: 이전 풀이를 보관하고 빈 파일에서 재도전
function psretry -d "재도전 시작"
    mv main.cpp main.(date +%m%d).cpp
    cp $PS_ROOT/template.cpp main.cpp
    date +%s > .start
    nvim main.cpp
end

# 컴파일 후 모든 *.in 을 실행해 *.out 과 비교
function psrun -d "빌드와 테스트"
    g++ -std=c++17 -O2 -g -Wall -Wextra -Wshadow \
        -fsanitize=address,undefined -D_GLIBCXX_DEBUG main.cpp -o sol
    or return 1
    set fail 0
    for f in *.in
        set exp (string replace -r '\.in$' '.out' $f)
        if timeout 2 ./sol < $f | diff -q - $exp >/dev/null
            echo "PASS $f"
        else
            echo "FAIL $f"
            timeout 2 ./sol < $f | diff - $exp
            set fail 1
        end
    end
    set now (date +%s)
    set start (cat .start)
    echo "경과: "(math -s0 "($now - $start) / 60")"분"
    return $fail
end

# psdone ok | fail | again [메모]
#   ok    : 제한 시간 안에 통과, 재도전 없음
#   fail  : 실패 또는 풀이 참고, 3일 뒤 재도전
#   again : 재도전에서 통과, 7일 뒤 최종 확인
function psdone -a result note -d "결과 기록"
    switch $result
        case ok;    set due -
        case fail;  set due (date -d "+3 days" +%F)
        case again; set due (date -d "+7 days" +%F)
        case '*';   echo "usage: psdone ok|fail|again [메모]"; return 1
    end
    set now (date +%s)
    set start (cat .start)
    set mins (math -s0 "($now - $start) / 60")
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' (date +%F) \
        (basename (dirname $PWD)) (basename $PWD) $result $mins $due "$note" \
        >> $PS_ROOT/log.tsv
end

# 문제별 마지막 기록 기준으로 오늘까지 도래한 재도전 목록
function psdue -d "오늘 재도전할 문제"
    awk -F'\t' -v t=(date +%F) '
        { key = $2 "/" $3; due[key] = $6; res[key] = $4 }
        END { for (k in due) if (due[k] != "-" && due[k] <= t) print due[k], k, res[k] }
    ' $PS_ROOT/log.tsv | sort
end
컴파일 옵션의 역할
옵션
잡아내는 것
-Wall -Wextra -Wshadow
부호 비교, 미사용 변수, 가려진 변수
-fsanitize=address
배열 범위 초과, 해제 후 사용
-fsanitize=undefined
부호 있는 정수 오버플로, 잘못된 시프트
-D_GLIBCXX_DEBUG
vector 인덱스 초과, 무효화된 iterator 사용
timeout 2
무한 루프와 명백한 시간 초과
sanitizer는 실행을 몇 배 느리게 하므로, 시간 초과 여부는 g++ -std=c++17 -O2 main.cpp -o fast로 따로 빌드해 최대 크기 입력으로 확인합니다. macOS 기본 컴파일러에는 bits/stdc++.h와 _GLIBCXX_DEBUG가 없으니 Homebrew gcc를 쓰거나 Linux에서 연습합니다.
LeetCode 문제를 로컬에서 푸는 형태
입력 형식은 직접 정합니다. 이 과정이 HackerRank식 입력 파싱 연습이 됩니다.
// 1.in:  4 9
//        2 7 11 15
// 1.out: 0 1
class Solution {
public:
    vector<int> twoSum(vector<int>& nums, int target) { /* ... */ }
};

int main() {
    int n, target; cin >> n >> target;
    vector<int> nums(n);
    for (auto& x : nums) cin >> x;
    auto ans = Solution().twoSum(nums, target);
    cout << ans[0] << ' ' << ans[1] << '\n';
}
neovim 설정
저장 후 분할 터미널에서 psrun을 실행하는 키 하나면 충분합니다.
vim.keymap.set('n', '<leader>r', '<cmd>w<cr><cmd>botright 12split | terminal fish -c psrun<cr>')
연습 디렉터리에서는 AI 자동완성과 스니펫 플러그인을 끕니다. 시험 에디터에는 기본 자동완성 정도만 있으므로, STL 함수 이름과 인자 순서는 손으로 칠 수 있어야 합니다.
하루 사용 흐름
psdue                              # 오늘 재도전할 문제 확인
cd ~/ps/lc/0056-merge-intervals; psretry
psrun; and psdone again            # 재도전 통과

psnew lc 0739 daily-temperatures   # 새 문제
psrun                              # 반복
psdone fail "단조 스택 떠올리지 못함"
9. 모의고사와 시험 당일
4주차에 90분 세트 모의고사를 최소 3회 치르고, 합격선은 "2문제 완주 + 나머지 부분 점수"로 잡습니다. 문항 수와 시간이 확인되면 그 값으로 바꿉니다.
세트 구성
회차
환경
구성
출처
1회
로컬 CLI
Easy 1 + Medium 2
풀지 않은 LeetCode 문제를 유형을 가리고 선택
2회
HackerRank 웹 에디터
Medium 3
Interview Preparation Kit의 풀지 않은 문제
3회
HackerRank 웹 에디터
Medium 2 + 구현형 1
위와 같음, 구현형은 문자열/시뮬레이션에서 선택
4회 (여유 시)
실제 플랫폼의 샘플 테스트
제공되는 그대로
응시 안내 메일의 링크
세트 규칙
1. 90분 동안 중단하지 않고, 검색과 풀이 참고 없이 풉니다.
2. 시작 후 5분 동안 모든 문제를 읽고 쉬운 순서로 풉니다.
3. 한 문제에 35분을 넘기면 부분 점수용 코드를 제출하고 넘어갑니다.
4. 종료 후 30분 안에 복기합니다: 문제별 소요 시간, 틀린 원인, 시간을 잃은 지점.
5. 못 푼 문제는 일반 문제와 똑같이 psdone fail로 기록해 재도전 대상에 넣습니다.
부분 점수 전략
숨김 테스트케이스는 보통 케이스별로 채점되므로, 최적 풀이가 떠오르지 않으면 느리더라도 정확한 풀이를 먼저 제출합니다. O(n^2) 풀이로 작은 케이스를 통과시킨 뒤 남는 시간에 개선하는 편이 빈 답안보다 낫습니다.
시험 전날
[ ] 응시 환경 확인: 브라우저, 네트워크, 웹캠과 화면 공유 요구 여부
[ ] 플랫폼 샘플 테스트 1회 완료
[ ] 4절의 함정 표와 6절의 복잡도 표 다시 읽기
[ ] 새 문제는 풀지 않고, 재도전 목록에서 2문제만 풀기
시험 당일 풀이 중 점검
[ ] 제약 조건을 읽고 목표 복잡도를 정했는가
[ ] 값 범위에서 long long이 필요한가
[ ] 빈 입력과 원소 1개에서 동작하는가
[ ] 출력 형식(공백, 개행, 대소문자)이 예제와 정확히 같은가
[ ] 디버그 출력을 지웠는가
[ ] 커스텀 입력으로 직접 만든 케이스를 한 번 돌렸는가
준비 완료 기준
항목
기준
필수 45문제
전부 한 번 이상 통과, 실패 문제는 재도전 통과
Medium 풀이 시간
풀어 본 유형 기준 25분 이내
세트 모의고사
최근 2회 연속 2문제 이상 완주
뼈대 코드
grid BFS, 이진탐색, 누적합, 백트래킹을 각 5분 안에 작성
