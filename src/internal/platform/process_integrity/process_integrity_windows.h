//go:build windows

#ifndef INTEGRITY_H
#define INTEGRITY_H

#ifdef __cplusplus
extern "C" {
#endif

unsigned int GetProcessLevel(unsigned int pid);

#ifdef __cplusplus
}
#endif

#endif // INTEGRITY_H
