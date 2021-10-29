package essay

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"strings"
)

type (
	Document interface {
		Note(args ...interface{})
		Section(name string, body interface{})
		Depth() int
	}

	Displayer interface {
		Display(Document)
	}

	Renderer interface {
		Render(Builtin) (interface{}, error)
	}

	Builtin interface {
		RenderImage(EncodedImage) (interface{}, error)
		RenderTable(Table) (interface{}, error)
	}

	Essay struct {
		config Config
		depth  int
		tmpl   *template.Template

		structuredDoc
	}

	Config struct {
		Dir   string
		Title string
	}

	structuredDoc struct {
		*Essay
		divs []interface{}
	}

	sectionRenderer struct {
		name string
		structuredDoc
	}

	noteRenderer struct {
		structuredDoc
	}

	displayRenderer struct {
		dtype string
		structuredDoc
	}

	funcDisplayer struct {
		docf func(Document)
	}
)

func New(conf Config) (*Essay, error) {
	e := &Essay{
		config: conf,
	}
	tmpl, err := template.New("essay").
		Funcs(map[string]interface{}{
			"css":     e.css,
			"body":    e.body,
			"section": e.section,
			"render":  e.render,
			"base64":  base64Encode,
			"indexof": indexOf,
		}).
		ParseGlob(templateGlobPath())
	if err != nil {
		return nil, err
	}
	e.Essay = e
	e.tmpl = tmpl
	return e, nil
}

func (e *Essay) Depth() int {
	return e.depth
}

func (e *Essay) Close() (err error) {
	if err = os.MkdirAll(e.config.Dir, os.ModePerm); err != nil {
		return
	}

	data, err := e.generate()
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(path.Join(e.config.Dir, "index.html"), []byte(data), os.ModePerm); err != nil {
		return
	}

	return nil
}

func (e *Essay) ascend() {
	e.depth--
}

func (e *Essay) descend() *Essay {
	e.depth++
	return e
}

func (e *Essay) generate() (template.HTML, error) {
	return e.execute("essay.html", struct {
		Heading string
		Divs    []interface{}
		Depth   int
	}{Heading: e.config.Title, Divs: e.divs, Depth: 1})
}

func (e *Essay) body(body []interface{}) (template.HTML, error) {
	return e.execute("body.html", body)
}

func (e *Essay) css(name string) (template.CSS, error) {
	css, err := e.execute(name, e.config)
	return template.CSS(css), err
}

func (e *Essay) section(section interface{}) (interface{}, error) {
	defer e.descend().ascend()
	return e.execute("section.html", section)
}

func (e *Essay) execute(name string, arg interface{}) (template.HTML, error) {
	var buf bytes.Buffer

	if err := e.Essay.tmpl.ExecuteTemplate(&buf, name, arg); err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}

func (doc *structuredDoc) Note(list ...interface{}) {
	note := &noteRenderer{}
	note.Essay = doc.Essay
	for _, something := range list {
		note.add(something)
	}
	doc.add(note)
}

func (doc *structuredDoc) Section(name string, body interface{}) {
	section := &sectionRenderer{
		name: name,
	}
	section.Essay = doc.Essay

	section.add(body)
	doc.add(section)
}

func (d *structuredDoc) add(c interface{}) {
	d.divs = append(d.divs, c)
}

func (e *Essay) render(arg interface{}) (interface{}, error) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println("Recovered panic", err)
	// 		return
	// 	}
	// }()
	switch t := arg.(type) {
	case string, template.HTML:
		return arg, nil
	case Renderer:
		return t.Render(e)
	case Displayer:
		return e.renderNamedDisplayer(t)
	case func(Document):
		return e.renderDisplayer("", funcDisplayer{t})
	default:
		return fmt.Sprintf("%v", arg), nil
	}
}

func (n *noteRenderer) Render(Builtin) (interface{}, error) {
	return n.execute("note.html", struct {
		Divs []interface{}
	}{Divs: n.divs})
}

func (s *sectionRenderer) Render(Builtin) (interface{}, error) {
	defer s.descend().ascend()
	return s.execute("section.html", struct {
		Heading string
		Depth   int
		Divs    []interface{}
	}{Heading: s.name, Depth: s.depth, Divs: s.divs})
}

func (d *displayRenderer) Render(Builtin) (interface{}, error) {
	return d.execute("display.html", struct {
		Type  string
		Depth int
		Divs  []interface{}
	}{Type: d.dtype, Depth: d.depth, Divs: d.divs})
}

func (e *Essay) renderNamedDisplayer(displayer Displayer) (interface{}, error) {
	defer e.descend().ascend()
	dtype := simplifyType(displayer)
	if stringer, ok := displayer.(fmt.Stringer); ok {
		dtype = dtype + ": " + stringer.String()
	}

	return e.renderDisplayer(dtype, displayer)
}

func (e *Essay) renderDisplayer(dtype string, displayer Displayer) (interface{}, error) {
	dd := &displayRenderer{
		dtype: dtype,
	}
	dd.Essay = e
	displayer.Display(dd)
	return dd.Render(e)
}

func simplifyType(d interface{}) string {
	v := reflect.TypeOf(d)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	s := v.String()
	if pos := strings.LastIndex(s, "."); pos >= 0 {
		s = s[pos+1:]
	}
	return s
}

func base64Encode(in []byte) template.HTML {
	return template.HTML(base64.StdEncoding.EncodeToString(in))
}

func indexOf(slice interface{}, idx int) interface{} {
	return reflect.ValueOf(slice).Index(idx).Interface()
}

func Main(title string, writer func(Document)) {
	ess, err := New(Config{
		Title: title,
		Dir:   strings.Replace(strings.ToLower(title), " ", "_", -1),
	})
	if err != nil {
		log.Fatal(err)
	}

	writer(ess)

	if err := ess.Close(); err != nil {
		log.Fatal(err)
	}
}

func (f funcDisplayer) Display(doc Document) {
	f.docf(doc)
}
