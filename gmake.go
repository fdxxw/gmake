package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	cfgFile string
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	var rootCmd = &cobra.Command{
		Use:   "gmake",
		Short: "parse custom makefile and execute",
		Long:  "",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ym := parseConfig(cfgFile)
			run(ym)
		},
	}
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "gmake.yml", "config file")
	rootCmd.Execute()
}

func parseConfig(cfgFile string) map[string]interface{} {
	ymlData, err := ioutil.ReadFile(cfgFile)
	checkError(err)
	m := make(map[string]interface{})
	err = yaml.Unmarshal(ymlData, &m)
	checkError(err)
	return m
}

func run(ym map[string]interface{}) {
	var vars map[interface{}]interface{}
	if v, ok := ym["vars"]; ok {
		vars = v.(map[interface{}]interface{})
	} else {
		vars = make(map[interface{}]interface{})
	}
	t := time.Now().Format("2006-01-02 15:04")
	vars["time"] = t
	cmdDir := ""
	for k, v := range ym {
		if k != "vars" {
			lines := strings.Split(cast.ToString(v), "\n")
			for _, line := range lines {
				if line != "" {
					// line = ResolveVars(vars, line)
					cmdStrs := strings.Fields(line)
					for i, cmdStr := range cmdStrs {
						cmdStrs[i] = ResolveVars(vars, cmdStr)
					}
					bin := cmdStrs[0]
					args := cmdStrs[1:]
					switch bin {
					case "@var":
						vars[args[0]] = strings.Join(args[1:], " ")
						break
					case "@env":
						os.Setenv(args[0], strings.Join(args[1:], " "))
						break
					case "#":
						break
					case "@echo":
						fmt.Println("@echo: " + strings.Join(args, " "))
						break
					case "@mv":
						mv(args[0], args[1])
						break
					case "@copy":
						copy(args[0], args[1])
						break
					case "@rm":
						rm(args[0])
						break
					case "@mkdir":
						mkdir(args[0])
						break
					case "@touch":
						touch(args[0])
						break
					case "@cd":
						abs, err := filepath.Abs(args[0])
						checkError(err)
						cmdDir = abs
						break
					default:
						cmd := exec.Command(bin, args...)
						if cmdDir != "" {
							cmd.Dir = cmdDir
						}
						ExecCmd(cmd)
					}
				}
			}
		}
	}
}

func mv(from, to string) {
	copy(from, to)
	rm(from)
}
func rm(path string) {
	err := os.RemoveAll(path)
	checkError(err)
}
func mkdir(path string) {
	err := os.MkdirAll(path, os.ModePerm)
	checkError(err)
}
func touch(path string) {
	f, err := os.Create(path)
	defer f.Close()
	checkError(err)
}
func copy(src, dst string) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	if isDir(src) {
		if isFile(dst) {
			panic(fmt.Errorf("不能复制目录到文件 src=%v dst=%v", src, dst))
		}
		si, err := os.Stat(src)
		checkError(err)
		// dst = path.Join(dst, filepath.Base(src))
		err = os.MkdirAll(dst, si.Mode())
		checkError(err)
		entries, err := ioutil.ReadDir(src)
		checkError(err)
		for _, entry := range entries {
			srcPath := filepath.Join(src, entry.Name())
			dstPath := filepath.Join(dst, entry.Name())

			if entry.IsDir() {
				copy(srcPath, dstPath)
			} else {
				// Skip symlinks.
				if entry.Mode()&os.ModeSymlink != 0 {
					continue
				}
				err = copyFile(srcPath, dstPath)
			}
		}
	} else {
		if isFile(dst) {
			copyFile(src, dst)
		} else {
			copyFile(src, path.Join(dst, filepath.Base(src)))
		}
	}
}
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func ResolveVars(vars interface{}, templateStr string) string {
	if vars == nil {
		return templateStr
	}
	t := template.Must(template.New("template").Parse(templateStr))
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, vars); err != nil {
		log.Fatal(err)
	}
	s := buf.String()
	return s
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return false
	}
	return s == nil || s.IsDir()
}
func isFile(path string) bool {
	s, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return false
	}
	return s == nil || !s.IsDir()
}
func ExecCmd(c *exec.Cmd) {
	log.Println(c.String())
	stdout, err := c.StdoutPipe()
	checkError(err)
	stderr, err := c.StderrPipe()
	checkError(err)
	err = c.Start()
	checkError(err)
	io.Copy(os.Stdout, stdout)
	io.Copy(os.Stderr, stderr)
	c.Wait()
}
