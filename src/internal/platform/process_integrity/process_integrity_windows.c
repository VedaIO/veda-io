//go:build windows

#include "process_integrity_windows.h"
#include <windows.h>
#include <securitybaseapi.h>
#include <stdlib.h>

unsigned int GetProcessLevel(unsigned int pid) {
    HANDLE hProcess = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, FALSE, pid);
    if (!hProcess) {
        return 0;
    }

    HANDLE hToken = NULL;
    if (!OpenProcessToken(hProcess, TOKEN_QUERY, &hToken)) {
        CloseHandle(hProcess);
        return 0;
    }

    DWORD tokenInfoLen = 0;
    GetTokenInformation(hToken, TokenIntegrityLevel, NULL, 0, &tokenInfoLen);
    
    unsigned int level = 0;
    
    if (tokenInfoLen > 0) {
        TOKEN_MANDATORY_LABEL* tokenInfo = (TOKEN_MANDATORY_LABEL*)malloc(tokenInfoLen);
        if (tokenInfo) {
            if (GetTokenInformation(hToken, TokenIntegrityLevel, tokenInfo, tokenInfoLen, &tokenInfoLen)) {
                DWORD subAuthCount = *GetSidSubAuthorityCount(tokenInfo->Label.Sid);
                if (subAuthCount > 0) {
                    level = *GetSidSubAuthority(tokenInfo->Label.Sid, subAuthCount - 1);
                }
            }
            free(tokenInfo);
        }
    }

    CloseHandle(hToken);
    CloseHandle(hProcess);

    return level;
}
