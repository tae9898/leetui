// null 포인터 쓰기 — sanitizer가 잡고, 비활성 환경에서는 SIGSEGV로 종료.
int main() {
    volatile int* p = nullptr;
    *p = 1;
    return 0;
}
