package main
import (
 "fmt"
 "os"
)
var version="v0.3.0"
func main() {
 if len(os.Args)<2 { fmt.Println("devkit <build|compose|init|list|validate|doctor|images|version>"); os.Exit(1)}
 switch os.Args[1] {
 case "version": fmt.Println(version)
 case "build": fmt.Println("build workflow scaffold")
 case "compose": fmt.Println("compose engine scaffold")
 case "doctor": fmt.Println("doctor: ok")
 case "images": fmt.Println("images: not implemented yet")
 default: fmt.Println("command:", os.Args[1])
 }
}
