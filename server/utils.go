package server

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/ije/gox/utils"
	"github.com/ije/gox/valid"
)

var (
	regFullVersion      = regexp.MustCompile(`^\d+\.\d+\.\d+[a-zA-Z0-9\.\+\-_]*$`)
	regFullVersionPath  = regexp.MustCompile(`([^/])@\d+\.\d+\.\d+[a-zA-Z0-9\.\+\-_]*/`)
	regBuildVersionPath = regexp.MustCompile(`^/v\d+/`)
	regLocPath      = regexp.MustCompile(`(\.[a-z]+):\d+:\d+$`)
	npmNaming           = valid.Validator{valid.FromTo{'a', 'z'}, valid.FromTo{'0', '9'}, valid.Eq('.'), valid.Eq('_'), valid.Eq('-')}
)

type stringSet struct {
	lock sync.RWMutex
	m    map[string]struct{}
}

func newStringSet() *stringSet {
	return &stringSet{m: map[string]struct{}{}}
}

func (s *stringSet) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.m)
}

func (s *stringSet) Has(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.m[key]
	return ok
}

func (s *stringSet) Add(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[key] = struct{}{}
}

func (s *stringSet) Values() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	a := make([]string, len(s.m))
	i := 0
	for key := range s.m {
		a[i] = key
		i++
	}
	return a
}

// sortable version slice
type versionSlice []string

func (s versionSlice) Len() int      { return len(s) }
func (s versionSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s versionSlice) Less(i, j int) bool {
	avs, aStage := utils.SplitByFirstByte(s[i], '-')
	bvs, bStage := utils.SplitByFirstByte(s[j], '-')
	av := strings.Split(avs, ".")
	bv := strings.Split(bvs, ".")
	if len(av) != 3 || len(bv) != 3 {
		return avs > bvs
	}
	if av[0] == bv[0] {
		if av[1] == bv[1] {
			if av[2] == bv[2] {
				return aStage > bStage
			}
			a2, _ := strconv.Atoi(av[2])
			b2, _ := strconv.Atoi(bv[2])
			return a2 > b2
		}
		a1, _ := strconv.Atoi(av[1])
		b1, _ := strconv.Atoi(bv[1])
		return a1 > b1
	}
	a0, _ := strconv.Atoi(av[0])
	b0, _ := strconv.Atoi(bv[0])
	return a0 > b0
}

func identify(importPath string) string {
	p := []byte(importPath)
	for i, c := range p {
		switch c {
		case '/', '-', '@', '.':
			p[i] = '_'
		default:
			p[i] = c
		}
	}
	return string(p)
}

func isRemoteImport(importPath string) bool {
	return strings.HasPrefix(importPath, "https://") || strings.HasPrefix(importPath, "http://")
}

func isLocalImport(importPath string) bool {
	return strings.HasPrefix(importPath, "file://") || strings.HasPrefix(importPath, "/") || strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") || importPath == "." || importPath == ".."
}

func startsWith(s string, prefixs ...string) bool {
	for _, prefix := range prefixs {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func endsWith(s string, suffixs ...string) bool {
	for _, suffix := range suffixs {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

type devFS struct {
	cwd string
}

func (fs devFS) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(fs.cwd, name))
}

func dirExists(filepath string) bool {
	fi, err := os.Lstat(filepath)
	return err == nil && fi.IsDir()
}

func fileExists(filepath string) bool {
	fi, err := os.Lstat(filepath)
	return err == nil && !fi.IsDir()
}

func ensureDir(dir string) (err error) {
	_, err = os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
	}
	return
}

func btoaUrl(s string) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(s)), "=")
}

func atobUrl(s string) (string, error) {
	if l := len(s) % 4; l > 0 {
		s += strings.Repeat("=", 4-l)
	}
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func kill(pidFile string) (err error) {
	data, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	return process.Kill()
}
