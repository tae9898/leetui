// 무한 루프 — volatile이므로 최적화로 제거되지 않는다.
int main() {
    volatile int x = 0;
    for (;;)
        ++x;
}
