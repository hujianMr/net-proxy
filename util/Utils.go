package util

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIfErrMsg(err error, msg string) {
	if err != nil {
		panic(msg)
	}
}

func CheckPort(port string) error {
	checkStatement := fmt.Sprintf(`netstat -anp | grep -q %d ; echo $?`, port)
	output, err := exec.Command("sh", "-c", checkStatement).CombinedOutput()
	if err != nil {
		return err
	}
	result, err := strconv.Atoi(strings.TrimSuffix(string(output), "\n"))
	if err != nil {
		return err
	}
	if result == 0 {
		return fmt.Errorf("端口%d已被占用", port)
	}

	return nil
}
