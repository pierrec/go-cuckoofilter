// This command is used to generate the Cuckoo filter types.
// Run "go generate" in the above directory.
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
)

type cData struct {
	Name, Desc string
	// NumBits: Number of bits in a bucket
	// NumBitsFingerprint: Number of bits in a fingerprint
	// MaxInserts: number of fingerprints per bucket
	NumBits, NumBitsFingerprint, NumBitsFingerprintUint, MaxInserts int
	// Maximum error rate
	ErrRate float64
}

// genFile generates a go file with the given name and using the supplied template and data.
func genFile(fname string, tmpl *template.Template, data *cData) {
	goFile, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer goFile.Close()
	err = tmpl.Execute(goFile, data)
	if err != nil {
		log.Fatal(err)
	}
}

// power2 gives the closest power of 2 for n.
func power2(n int) int {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

func main() {
	// Load the templates.
	tmpl, err := template.ParseGlob("cuckoo.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmplTest, err := template.ParseGlob("cuckoo_test.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmplBench, err := template.ParseGlob("bench_test.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	// Generate Cuckoo filter types.
	for _, data := range []cData{
		{Name: "S", NumBits: 8, NumBitsFingerprint: 4, Desc: "1 byte per item", ErrRate: 11, MaxInserts: 2},
		{Name: "M", NumBits: 16, NumBitsFingerprint: 8, Desc: "2 bytes per item", ErrRate: 1, MaxInserts: 2},
		{Name: "L", NumBits: 32, NumBitsFingerprint: 16, Desc: "4 bytes per item", ErrRate: 0.005, MaxInserts: 2},
		// similar error rate as bucket 32 / fingerprint 16 with twice the memory footprint
		// {Name: "L", NumBits: 64, NumBitsFingerprint: 16, Desc: "8 bytes per item", ErrRate: 0.005, MaxInserts: 4},
	} {
		log.Printf("Generating %s\n", data.Name)
		// Compute the uint size to hold a fingerprint
		data.NumBitsFingerprintUint = power2(data.NumBitsFingerprint)
		if data.NumBitsFingerprintUint < 8 {
			data.NumBitsFingerprintUint = 8
		}
		genFile(fmt.Sprintf("cuckoo%s.go", data.Name), tmpl, &data)
		genFile(fmt.Sprintf("cuckoo%s_test.go", data.Name), tmplTest, &data)
		genFile(fmt.Sprintf("bench%s_test.go", data.Name), tmplBench, &data)
	}
	log.Println("Running go fmt")
	cmd := exec.Command("go", "fmt")
	cmd.Dir = "."
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
