package assets

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type file struct {
	info os.FileInfo
	path string
}

// An asset generator. The generator can be used to generate an asset go file
// with all the assets that were added to the generator embedded into it.
// The generated assets are made available by the specified go variable
// VariableName which is of type assets.FileSystem.
type Generator struct {
	// The package name to generate assets in,
	PackageName string

	// The variable name containing the asset filesystem (defaults to Assets),
	VariableName string

	// Strip the specified prefix from all paths,
	StripPrefix string

	fsDirsMap  map[string][]string
	fsFilesMap map[string]file
}

func (x *Generator) addPath(parent string, prefix string, info os.FileInfo) error {
	p := path.Join(parent, info.Name())

	f := file{
		info: info,
		path: path.Join(prefix, p),
	}

	x.fsFilesMap[p] = f

	if info.IsDir() {
		f, err := os.Open(f.path)
		fi, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			return err
		}

		x.fsDirsMap[p] = make([]string, 0, len(fi))

		for _, f := range fi {
			if err := x.addPath(p, prefix, f); err != nil {
				return err
			}
		}
	} else {
		x.appendFileInDir(parent, info.Name())
	}

	return nil
}

func (x *Generator) appendFileInDir(dir string, file string) {
	for _, v := range x.fsDirsMap[dir] {
		if v == file {
			return
		}
	}

	x.fsDirsMap[dir] = append(x.fsDirsMap[dir], file)
}

func (x *Generator) addParents(p string, prefix string) error {
	dname, fname := path.Split(p)

	if len(dname) == 0 {
		return nil
	}

	wosep := dname[0 : len(dname)-1]

	if err := x.addParents(wosep, prefix); err != nil {
		return err
	}

	if len(wosep) == 0 {
		wosep = "/"
	}

	x.appendFileInDir(wosep, fname)

	if _, ok := x.fsFilesMap[wosep]; !ok {
		pp := path.Join(prefix, wosep)
		s, err := os.Stat(pp)

		if err != nil {
			return err
		}

		x.fsFilesMap[wosep] = file{
			info: s,
			path: pp,
		}
	}

	return nil
}

func (x *Generator) splitRelPrefix(p string) (string, string) {
	i := 0
	relp := "../"

	for strings.HasPrefix(p[i:], relp) {
		i += len(relp)
	}

	return path.Join(p[0:i], "."), path.Join("/", p[i:])
}

// Add a file or directory asset to the generator. Added directories will be
// recursed automatically.
func (x *Generator) Add(p string) error {
	if x.fsFilesMap == nil {
		x.fsFilesMap = make(map[string]file)
	}

	if x.fsDirsMap == nil {
		x.fsDirsMap = make(map[string][]string)
	}

	p = path.Clean(p)

	info, err := os.Stat(p)

	if err != nil {
		return err
	}

	prefix, p := x.splitRelPrefix(p)

	if err := x.addParents(p, prefix); err != nil {
		return err
	}

	return x.addPath(path.Dir(p), prefix, info)
}

func (x *Generator) stripPrefix(p string) (string, bool) {
	if len(x.StripPrefix) == 0 {
		return p, true
	}

	if strings.HasPrefix(p, x.StripPrefix) {
		return p[len(x.StripPrefix):], true
	} else {
		return p, false
	}
}

// Write the asset tree specified in the generator to the given writer. The
// written asset tree is a valid, standalone go file with the assets
// embedded into it.
func (x *Generator) Write(wr io.Writer) error {
	p := x.PackageName

	if len(p) == 0 {
		p = "main"
	}

	variableName := x.VariableName

	if len(variableName) == 0 {
		variableName = "Assets"
	}

	writer := &bytes.Buffer{}

	// Write package and import
	fmt.Fprintf(writer, "package %s\n\n", p)
	fmt.Fprintln(writer, "import (")
	fmt.Fprintln(writer, "\t\"time\"")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "\t\"github.com/jessevdk/go-assets\"")
	fmt.Fprintln(writer, ")")
	fmt.Fprintln(writer)

	vnames := make(map[string]string)

	// Write file contents as const strings
	if x.fsFilesMap != nil {
		// Create mapping from full file path to asset variable name.
		// This also reads the file and writes the contents as a const
		// string
		for k, v := range x.fsFilesMap {
			if v.info.IsDir() {
				continue
			}

			f, err := os.Open(v.path)

			if err != nil {
				return err
			}

			data, err := ioutil.ReadAll(f)

			f.Close()

			if err != nil {
				return err
			}

			s := sha1.New()
			io.WriteString(s, k)

			vname := fmt.Sprintf("_%s%x", variableName, s.Sum(nil))
			vnames[k] = vname

			fmt.Fprintf(writer, "var %s = %#v\n", vname, string(data))
		}

		fmt.Fprintln(writer)
	}

	fmt.Fprintf(writer, "// %s returns go-assets FileSystem\n", variableName)
	fmt.Fprintf(writer, "var %s = assets.NewFileSystem(", variableName)

	if x.fsDirsMap == nil {
		x.fsDirsMap = make(map[string][]string)
	}

	if x.fsFilesMap == nil {
		x.fsFilesMap = make(map[string]file)
	}

	dirmap := make(map[string][]string)

	for k, v := range x.fsDirsMap {
		if kk, ok := x.stripPrefix(k); ok {
			if len(kk) == 0 {
				kk = "/"
			}

			dirmap[kk] = v
		}
	}

	fmt.Fprintf(writer, "%#v, ", dirmap)
	fmt.Fprintf(writer, "map[string]*assets.File{\n")

	// Write files
	for k, v := range x.fsFilesMap {
		kk, ok := x.stripPrefix(k)

		if !ok {
			continue
		}

		if len(kk) == 0 {
			kk = "/"
		}

		mt := v.info.ModTime()

		var dt string

		if !v.info.IsDir() {
			dt = "[]byte(" + vnames[k] + ")"
		} else {
			dt = "nil"
		}

		fmt.Fprintf(writer, "\t\t%#v: &assets.File{\n", kk)
		fmt.Fprintf(writer, "\t\t\tPath: %#v,\n", kk)
		fmt.Fprintf(writer, "\t\t\tFileMode: %#v,\n", v.info.Mode())
		fmt.Fprintf(writer, "\t\t\tMtime: time.Unix(%#v, %#v),\n", mt.Unix(), mt.UnixNano())
		fmt.Fprintf(writer, "\t\t\tData: %s,\n", dt)
		fmt.Fprintf(writer, "\t\t},")
	}

	fmt.Fprintln(writer, "\t}, \"\")")

	ret, err := format.Source(writer.Bytes())

	if err != nil {
		return err
	}

	wr.Write(ret)
	return nil
}
