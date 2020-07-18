package main

import (
	"fmt"
	"io"
	"os"
	//"path/filepath"
	//"strings"
	"sort"
)

func dirTree(out io.Writer, path string, printFiles bool) error {
	current_file, err := os.Open(path)
	if err != nil {
		return err
	}
	res, err1 := current_file.Readdir(-1) // смотрим, что в директории
	
	sort.Sort(ByName(res)) // сортируем

	if !printFiles { // при необходимости исключаем файлы
		res = filterDirs(res)
	}
	for index, element := range res {
		if element.IsDir() {
			if index == len(res) - 1 {
				fmt.Fprintf(out, "└───%v\n", element.Name())
				printFolder(out, path + "/" + element.Name(), printFiles, "")
			} else {
				fmt.Fprintf(out, "├───%v\n", element.Name())
				printFolder(out, path + "/" + element.Name(), printFiles, "│")
			}
		} else if printFiles {
			if element.Size() == 0 {
				if index == len(res) - 1 {
					fmt.Fprintf(out, "└───%v (empty)\n", element.Name())
				} else {
					fmt.Fprintf(out, "├───%v (empty)\n", element.Name())
				}
			} else {
				if index == len(res) - 1 {
					fmt.Fprintf(out, "└───%v (%vb)\n", element.Name(), element.Size())
				} else {
					fmt.Fprintf(out, "├───%v (%vb)\n", element.Name(), element.Size())
				}
			}
		}
	}

	return err1
}

func printFolder(out io.Writer, path string, printFiles bool, outputPrefix string) error {
	current_file, err := os.Open(path)
	if err != nil {
		return err
	}
	res, err1 := current_file.Readdir(-1) // смотрим, что в директории
	
	sort.Sort(ByName(res)) // сортируем

	if !printFiles { // при необходимости исключаем файлы
		res = filterDirs(res)
	}
	for index, element := range res {
		if element.IsDir() {
			if index == len(res) - 1 {
				fmt.Fprintf(out, "%v\t└───%v\n", outputPrefix, element.Name())
				printFolder(out, path + "/" + element.Name(), printFiles, outputPrefix + "\t")
			} else {
				fmt.Fprintf(out, "%v\t├───%v\n", outputPrefix, element.Name())
				printFolder(out, path + "/" + element.Name(), printFiles, outputPrefix + "\t│")
			}
		} else if printFiles {
			if element.Size() == 0 {
				if index == len(res) - 1 {
					fmt.Fprintf(out, "%v\t└───%v (empty)\n", outputPrefix, element.Name())
				} else {
					fmt.Fprintf(out, "%v\t├───%v (empty)\n", outputPrefix, element.Name())
				}
			} else {
				if index == len(res) - 1 {
					fmt.Fprintf(out, "%v\t└───%v (%vb)\n", outputPrefix, element.Name(), element.Size())
				} else {
					fmt.Fprintf(out, "%v\t├───%v (%vb)\n", outputPrefix, element.Name(), element.Size())
				}
			}
		}
	}

	return err1
}

// функция, возвращающая только дикертории (исключает из массива файлы)
func filterDirs(files []os.FileInfo) []os.FileInfo {
    dirs := []os.FileInfo{}
    for _, value := range files {
        if(value.IsDir()) {
            dirs = append(dirs, value)
        }
    }
    return dirs
}

// для сортировки
type ByName []os.FileInfo

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }

func main() {
	out := os.Stdout // куда вывод
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
