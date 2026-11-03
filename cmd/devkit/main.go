package main
import (
 "fmt"
 "os"
)
var version="v0.2.0"
func main(){
 if len(os.Args)<2 {fmt.Println("devkit <init|list|validate|compose|doctor|images|version>"); os.Exit(1)}
 switch os.Args[1] {
 case "version": fmt.Println(version)
 case "doctor": fmt.Println("doctor: ok")
 case "images": fmt.Println("images: not implemented yet")
 default: fmt.Println("v0.2 command:", os.Args[1])
 }
}
