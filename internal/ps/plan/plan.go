// Package plan은 plan.md 5절의 문제 목록을 내장한다.
// 목록은 고정 스펙이므로 코드로 유지한다 (plan.md 파싱 대신).
package plan

type Item struct {
	Week  int    // 1~4, 0=여유
	Type  string // 유형
	Site  string // lc | hr | task
	ID    string // 디렉터리 접두 번호
	Name  string // 슬러그
	Title string // 표시용 제목
	Note  string // 핵심
}

var all = []Item{
	// 1주차 — 워밍업 Easy 16
	{1, "해시", "lc", "0001", "two-sum", "Two Sum", "unordered_map"},
	{1, "해시", "lc", "0217", "contains-duplicate", "Contains Duplicate", "unordered_set"},
	{1, "해시", "lc", "0242", "valid-anagram", "Valid Anagram", "배열 카운팅"},
	{1, "투포인터", "lc", "0125", "valid-palindrome", "Valid Palindrome", "문자 판별"},
	{1, "이진탐색", "lc", "0704", "binary-search", "Binary Search", "경계 처리"},
	{1, "배열", "lc", "0121", "best-time-to-buy-and-sell-stock", "Best Time to Buy and Sell Stock", "한 번 스캔"},
	{1, "스택", "lc", "0020", "valid-parentheses", "Valid Parentheses", "stack"},
	{1, "배열", "lc", "0088", "merge-sorted-array", "Merge Sorted Array", "뒤에서 병합"},
	{1, "배열", "lc", "0283", "move-zeroes", "Move Zeroes", "in-place"},
	{1, "비트", "lc", "0136", "single-number", "Single Number", "XOR"},
	{1, "배열", "lc", "0118", "pascals-triangle", "Pascal's Triangle", "2차원 vector"},
	{1, "DP", "lc", "0053", "maximum-subarray", "Maximum Subarray", "카데인"},
	{1, "연결리스트", "lc", "0206", "reverse-linked-list", "Reverse Linked List", "포인터 조작"},
	{1, "연결리스트", "lc", "0021", "merge-two-sorted-lists", "Merge Two Sorted Lists", "더미 노드"},
	{1, "연결리스트", "lc", "0141", "linked-list-cycle", "Linked List Cycle", "빠른/느린 포인터"},
	{1, "트리", "lc", "0104", "maximum-depth-of-binary-tree", "Maximum Depth of Binary Tree", "재귀 DFS"},

	// 2주차 — Medium 전반 13
	{2, "해시", "lc", "0049", "group-anagrams", "Group Anagrams", "정렬 키"},
	{2, "해시", "lc", "0347", "top-k-frequent-elements", "Top K Frequent Elements", "카운팅+힙"},
	{2, "해시", "lc", "0560", "subarray-sum-equals-k", "Subarray Sum Equals K", "누적합+해시"},
	{2, "투포인터", "lc", "0003", "longest-substring-without-repeating-characters", "Longest Substring Without Repeating", "슬라이딩"},
	{2, "투포인터", "lc", "0011", "container-with-most-water", "Container With Most Water", "양끝 좁히기"},
	{2, "투포인터", "lc", "0015", "3sum", "3Sum", "중복 건너뛰기"},
	{2, "배열", "lc", "0238", "product-of-array-except-self", "Product of Array Except Self", "앞뒤 누적곱"},
	{2, "배열", "lc", "0055", "jump-game", "Jump Game", "그리디"},
	{2, "구간", "lc", "0056", "merge-intervals", "Merge Intervals", "정렬 후 병합"},
	{2, "스택", "lc", "0739", "daily-temperatures", "Daily Temperatures", "단조 스택"},
	{2, "스택", "lc", "0150", "evaluate-reverse-polish-notation", "Evaluate RPN", "스택 계산"},
	{2, "이진탐색", "lc", "0153", "find-minimum-in-rotated-sorted-array", "Find Min in Rotated Array", "조건 분기"},
	{2, "이진탐색", "lc", "0875", "koko-eating-bananas", "Koko Eating Bananas", "파라메트릭"},

	// 3주차 — Medium 후반 16 + 과제형 3
	{3, "BFS/DFS", "lc", "0200", "number-of-islands", "Number of Islands", "grid 뼈대"},
	{3, "BFS/DFS", "lc", "0994", "rotting-oranges", "Rotting Oranges", "다중 시작점"},
	{3, "BFS/DFS", "lc", "0102", "binary-tree-level-order-traversal", "Level Order Traversal", "레벨 단위"},
	{3, "그래프", "lc", "0207", "course-schedule", "Course Schedule", "위상정렬"},
	{3, "백트래킹", "lc", "0078", "subsets", "Subsets", "선택/비선택"},
	{3, "백트래킹", "lc", "0046", "permutations", "Permutations", "방문 배열"},
	{3, "DP", "lc", "0198", "house-robber", "House Robber", "1차원 점화식"},
	{3, "DP", "lc", "0322", "coin-change", "Coin Change", "최솟값 DP"},
	{3, "비트", "lc", "0191", "number-of-1-bits", "Number of 1 Bits", "비트 순회"},
	{3, "비트", "lc", "0190", "reverse-bits", "Reverse Bits", "시프트·마스크"},
	{3, "비트", "lc", "0338", "counting-bits", "Counting Bits", "비트+DP"},
	{3, "구현", "lc", "0008", "string-to-integer-atoi", "String to Integer (atoi)", "오버플로"},
	{3, "구현", "lc", "0394", "decode-string", "Decode String", "스택 파서"},
	{3, "구현", "lc", "0054", "spiral-matrix", "Spiral Matrix", "경계 관리"},
	{3, "설계", "lc", "0146", "lru-cache", "LRU Cache", "list+map"},
	{3, "힙", "lc", "0973", "k-closest-points-to-origin", "K Closest Points", "크기 K 힙"},
	{3, "과제", "task", "0301", "ring-buffer", "링버퍼", "full/empty 구분"},
	{3, "과제", "task", "0302", "memory-pool", "고정 블록 메모리 풀", "free list"},
	{3, "과제", "task", "0303", "protocol-parser", "프로토콜 파서", "상태머신"},

	// 4주차 — HackerRank 적응 13
	{4, "적응", "hr", "0401", "sales-by-match", "Sales by Match", "영어 지문"},
	{4, "적응", "hr", "0402", "counting-valleys", "Counting Valleys", "영어 지문"},
	{4, "적응", "hr", "0403", "jumping-on-the-clouds", "Jumping on the Clouds", "영어 지문"},
	{4, "적응", "hr", "0404", "repeated-string", "Repeated String", "영어 지문"},
	{4, "배열", "hr", "0405", "2d-array-ds", "2D Array - DS", "hourglass"},
	{4, "배열", "hr", "0406", "arrays-left-rotation", "Arrays: Left Rotation", "회전"},
	{4, "배열", "hr", "0407", "new-year-chaos", "New Year Chaos", " inversion"},
	{4, "배열", "hr", "0408", "minimum-swaps-2", "Minimum Swaps 2", "사이클 분할"},
	{4, "문자열/스택", "hr", "0409", "sherlock-and-anagrams", "Sherlock and Anagrams", "anagram 쌍"},
	{4, "문자열/스택", "hr", "0410", "balanced-brackets", "Balanced Brackets", "스택"},
	{4, "그래프/DP", "hr", "0411", "bfs-shortest-reach", "BFS: Shortest Reach", "인접리스트"},
	{4, "그래프/DP", "hr", "0412", "dfs-connected-cell", "DFS: Connected Cell", "grid DFS"},
	{4, "그래프/DP", "hr", "0413", "max-array-sum", "Max Array Sum", "DP"},

	// 여유가 있으면
	{0, "이진탐색", "lc", "0033", "search-in-rotated-sorted-array", "Search in Rotated Sorted Array", "회전 탐색"},
	{0, "DP", "lc", "0300", "longest-increasing-subsequence", "Longest Increasing Subsequence", "LIS"},
	{0, "DP", "lc", "0062", "unique-paths", "Unique Paths", "격자 DP"},
	{0, "백트래킹", "lc", "0079", "word-search", "Word Search", "grid 백트래킹"},
	{0, "그래프", "lc", "0743", "network-delay-time", "Network Delay Time", "다익스트라"},
}

// All은 계획 순서 그대로 반환한다.
func All() []Item { return all }

// WeekLabel은 주차 표기.
func WeekLabel(w int) string {
	switch w {
	case 0:
		return "여유"
	default:
		return string(rune('0'+w)) + "주"
	}
}
