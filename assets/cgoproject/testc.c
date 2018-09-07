#include "testc.h"
#include <_cgo_export.h>
#include <string.h>

extern void HollyShit();
extern void PrintText(GoString text);
int MyPrint(char * txt)
{
    printf("in MyPrint: %s", txt);
    HollyShit();
    _GoString_ str;
    str.p = "love urself always.";
    str.n = strlen(str.p);
 //   PrintText(str);
    return 0;
}
