//go:build windows

#ifndef WINDOW_H
#define WINDOW_H

#ifdef __cplusplus
extern "C" {
#endif

int HasVisibleWindow(unsigned int pid);

#ifdef __cplusplus
}
#endif

#endif // WINDOW_H
