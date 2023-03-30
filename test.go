package main

// #cgo pkg-config: python3
// #cgo CFLAGS: -I/home/tone/.pyenv/versions/3.8.16/include/python3.8
// #cgo LDFLAGS: -L/home/tone/.pyenv/versions/3.8.16/lib/python3.8/config-3.8-x86_64-linux-gnu -lpython3.8
// #include <Python.h>
import "C"
import "unsafe"

// import "github.com/DataDog/go-python3"

func main() {

	pycodeGo := "print('hello world')"

	defer C.Py_Finalize()
	C.Py_Initialize()
	pycodeC := C.CString(pycodeGo)
	defer C.free(unsafe.Pointer(pycodeC))
	C.PyRun_SimpleString(pycodeC)
	// python3.PyRun_SimpleString("print('hello world')")

}
