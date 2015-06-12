#include <stdio.h>

int printi(int i) {
    printf("%d\n", i);
    return 0;
}

int readi() {
    int i;
    scanf("%d", &i);
    return i;
}

double printd(double d) {
    printf("%f\n", d);
    return 0;
}

double readd() {
    double d;
    scanf("%lf", &d);
    return d;
}
