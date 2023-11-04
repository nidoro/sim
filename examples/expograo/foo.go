package main

import (
    "fmt"
    "log"
    "os"
    "encoding/csv"
    "strconv"
    "github.com/RyanCarrier/dijkstra"
)

func Check(err error) {
    if err != nil {
        log.Fatalf("ERROR: %s", err)
    }
}

func GetIndex(ids []string, needle string) int {
    for i, id := range ids {
        if needle == id {
            return i
        }
    }
    return -1
}

func main() {
    file, err := os.Open("data/segmentos-RMS.tsv")
    Check(err)
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comma = '\t'
    records, err := reader.ReadAll()
    Check(err)
    
    matrix := make(map[string]map[string]float64)
    
    for _, row := range records[1:len(records)] {
        a, ok := matrix[row[0]]
        if !ok {
            a = make(map[string]float64)
            matrix[row[0]] = a
        }
        
        a[row[1]], _ = strconv.ParseFloat(row[2], 64)
        
        b, ok := matrix[row[1]]
        if !ok {
            b = make(map[string]float64)
            matrix[row[1]] = b
        }
        
        b[row[0]], _ = strconv.ParseFloat(row[2], 64)
        
        a[row[0]] = 0
        b[row[1]] = 0
    }
    
    ids := []string{}
    for id, _ := range matrix {
        //fmt.Printf("\t%s", id)
        ids = append(ids, id)
    }
    
    /*
    fmt.Println()
    
    for _, id1 := range ids {
        fmt.Printf("%s", id1)
        
        for _, id2 := range ids {
             fmt.Printf("\t%.4f", matrix[id1][id2])
        }
        
        fmt.Println()
    }
    */
    
    graph := dijkstra.NewGraph()
    
    for i, _ := range ids {
        graph.AddVertex(i)
    }
    
    for i, id1 := range ids {
        for j, id2 := range ids {
            if matrix[id1][id2] > 0.0 {
                graph.AddArc(i, j, int64(matrix[id1][id2] * 1000))
            }
        }
    }
    
    best, err := graph.Shortest(GetIndex(ids, "Apucarana (LAP), km 581,775"), GetIndex(ids, "California do Sul (LCF), km 560,730"))
    if err!=nil{
        log.Fatal(err)
    }
    fmt.Println("Shortest distance ", best.Distance, " following path ", best.Path)
}

