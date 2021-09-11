.386
.model flat, stdcall
option casemap :none

__UNICODE__     equ     1

include mem_alloc.inc

include /masm32/include/kernel32.inc
include /masm32/include/user32.inc
include /masm32/include/masm32rt.inc

includelib /masm32/lib/kernel32.lib
includelib /masm32/lib/user32.lib


.code
AllocMemory     proc    data_size: DWORD
    invoke  VirtualAlloc, NULL, data_size, MEM_COMMIT or MEM_RESERVE, PAGE_READWRITE
    .if     eax == NULL
        print   "Failed to allocate memory.", 0Dh, 0Ah
        invoke  ExitProcess, EXIT_FAILURE
    .endif

    push    eax
    invoke  RtlZeroMemory, eax, data_size
    pop     eax
    ret
AllocMemory     endp


FreeMemory      proc    data: ptr BYTE
    .if data != NULL
        invoke  VirtualFree, data, 0, MEM_RELEASE
        mov     data, NULL
    .endif
    ret
FreeMemory      endp

end