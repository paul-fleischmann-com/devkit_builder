package main
import (
 "fmt"
 "os"
)
func main(){
 if len(os.Args)<2 {fmt.Println("devkit <build|compose|init|validate|doctor>"); os.Exit(1)}
 fmt.Println("v2 skeleton command:",os.Args[1])
}
