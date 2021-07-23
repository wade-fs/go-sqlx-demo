- 整合下列兩個:
1. github: [mapstructure](https://github.com/mitchellh/mapstructure)
1. github: [structs](https://github.com/fatih/structs)
- 基本用法:
```go
package main
import (
    "fmt"
    ms "q8client/utils/mapstruct"
)

type Person struct {
    Name   string
    Age    int
    Emails []string
    Extra  map[string]string
}

func main() {
    // map 中的 key 不限大小寫, 但是上面 Person struct 則必須大寫開頭
    input := map[string]interface{}{
        "Name":   "Mitchell",
        "age":    91,
        "emails": []string{"one", "two", "three"},
        "Extra": map[string]string{
            "twitter": "mitchellh",
        },
    }
    fmt.Printf("input map:\n\t%#v\n", input)
    
    var result Person
    err := ms.Decode(input, &result)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Person struct:\n\t%#v\n", result)

    m := ms.Map(&result) // convert struct result Person to map
    fmt.Printf("convert struct to map:\n\t%#v\n", m)
}
```
