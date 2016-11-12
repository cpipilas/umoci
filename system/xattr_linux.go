/*
 * umoci: Umoci Modifies Open Containers' Images
 * Copyright (C) 2016 SUSE LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package system

import (
	"bytes"
	"fmt"
	"syscall"
	"unsafe"
)

// Llistxattr is a wrapper around llistxattr(2).
func Llistxattr(path string) ([]string, error) {
	bufsize, _, err := syscall.RawSyscall(syscall.SYS_LLISTXATTR, //. int llistxattr(
		uintptr(assertPtrFromString(path)), // char *path,
		0, // char *list,
		0) // size_t size);
	if err != 0 {
		return nil, fmt.Errorf("llistxattr: getting bufsize: %s", err)
	}

	if bufsize == 0 {
		return []string{}, nil
	}

	buffer := make([]byte, bufsize)
	n, _, err := syscall.RawSyscall(syscall.SYS_LLISTXATTR, // int llistxattr(
		uintptr(assertPtrFromString(path)),  // char *path,
		uintptr(unsafe.Pointer(&buffer[0])), // char *list,
		uintptr(bufsize))                    // size_t size);
	if err == syscall.ERANGE || n != bufsize {
		return nil, fmt.Errorf("llistxattr: getting buffer: xattr set changed")
	} else if err != 0 {
		return nil, fmt.Errorf("llistxattr: getting buffer: %s", err)
	}

	var xattrs []string
	for _, name := range bytes.Split(buffer, []byte{'\x00'}) {
		xattrs = append(xattrs, string(name))
	}
	return xattrs, nil
}

// Lremovexattr is a wrapper around lremovexattr(2).
func Lremovexattr(path, name string) error {
	_, _, err := syscall.RawSyscall(syscall.SYS_LREMOVEXATTR, // int lremovexattr(
		uintptr(assertPtrFromString(path)), //.   char *path
		uintptr(assertPtrFromString(name)), //.   char *name);
		0)
	if err != 0 {
		return fmt.Errorf("lremovexattr(%s, %s): %s", path, name, err)
	}
	return nil
}

// Lsetxattr is a wrapper around lsetxattr(2).
func Lsetxattr(path, name string, value []byte, flags int) error {
	_, _, err := syscall.RawSyscall6(syscall.SYS_LSETXATTR, // int lsetxattr(
		uintptr(assertPtrFromString(path)), //.   char *path,
		uintptr(assertPtrFromString(name)), //.   char *name,
		uintptr(unsafe.Pointer(&value[0])), //.   void *value,
		uintptr(len(value)),                //.   size_t size,
		uintptr(flags),                     //.   int flags);
		0)
	if err != 0 {
		return fmt.Errorf("lsetxattr(%s, %s, %s, %d): %s", path, name, value, flags, err)
	}
	return nil
}

// Lgetxattr is a wrapper around lgetxattr(2).
func Lgetxattr(path string, name string) ([]byte, error) {
	bufsize, _, err := syscall.RawSyscall6(syscall.SYS_LGETXATTR, //. int lgetxattr(
		uintptr(assertPtrFromString(path)), // char *path,
		uintptr(assertPtrFromString(name)), // char *name,
		0, // void *value,
		0, // size_t size);
		0, 0)
	if err != 0 {
		return nil, fmt.Errorf("lgetxattr: getting bufsize: %s", err)
	}

	if bufsize == 0 {
		return []byte{}, nil
	}

	buffer := make([]byte, bufsize)
	n, _, err := syscall.RawSyscall6(syscall.SYS_LGETXATTR, // int lgetxattr(
		uintptr(assertPtrFromString(path)),  // char *path,
		uintptr(assertPtrFromString(name)),  // char *name,
		uintptr(unsafe.Pointer(&buffer[0])), // void *value,
		uintptr(bufsize),                    // size_t size);
		0, 0)
	if err == syscall.ERANGE || n != bufsize {
		return nil, fmt.Errorf("lgetxattr: getting buffer: xattr set changed")
	} else if err != 0 {
		return nil, fmt.Errorf("lgetxattr: getting buffer: %s", err)
	}

	return buffer, nil
}

// Lclearxattrs is a wrapper around Llistxattr and Lremovexattr, which attempts
// to remove all xattrs from a given file.
func Lclearxattrs(path string) error {
	names, err := Llistxattr(path)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := Lremovexattr(path, name); err != nil {
			return err
		}
	}
	return nil
}