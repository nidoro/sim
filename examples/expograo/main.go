package main

import (
    "github.com/nidoro/sim"
    "os"
    "fmt"
    "log"
    "strconv"
    "encoding/csv"
    "math"
    //"github.com/kr/pretty"
)

var NumDays [12]int = [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
var MonthName [12]string = [12]string{
    "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dec",
}

const (
    Minutes float64 = 60
    Hours float64 = 60*Minutes
    Days float64 = 24*Hours
    Years float64 = 365*Days
    
    KTon = 1000
)

type Commodity struct {
    Id              string
    AnnualExports    float64
}

// Terminal Activities:
// WeighingIn       Normal(1.2, 0.1)
// Reception        Lognormal(1.9, 0.7)
// Classification   Triangular(3, 0.8)
// Unloading        Lognormal(4.3, 0.6)
// WeighingOut      Normal(1.2, 0.1)

type Terminal struct {
    Id                      string
    MonthExports            map[string]*[12]float64
    AnnualExports           map[string]float64
    Sazonality              map[string]*[12]float64
    NumTrucks               map[string]*[12]int
    Storage                 float64
    ProcessingRate          float64
    
    // simulated
    Exports                 map[string]*[12]float64
}

type CommodityProductivityInHarbor struct {
    DWT             float64
    OperatingTime   float64
    IdleTime        float64
    DockingTime     float64
    Productivity    float64
}

type Harbor struct {
    Id                      string
    MonthExports            map[string]*[12]float64
    AnnualExports           map[string]float64
    Sazonality              map[string]*[12]float64
    NumShips                map[string]*[12]int
    Storage                 float64
    Docks                   float64
    DockingInterval         float64
    ProcessingRate          float64
    Productivity            map[string]*CommodityProductivityInHarbor
    ShipDWTRNG              sim.RNGDiscrete
    
    // simulated
    Exports                 map[string]*[12]float64
}

type Truck struct {
    sim.EntityBase
    CommodityId string
    TerminalId  string
    Load        float64
}

type Train struct {
    sim.EntityBase
    CommodityId string
    TerminalId  string
    Load        float64
}

type Ship struct {
    sim.EntityBase
    CommodityId string
    HarborId    string
    DWT         float64
}

func ForwardToReception(entity sim.Entity) {
    env := entity.GetEnvironment()
    truck := sim.Cast[*Truck](entity)
    env.ForwardTo(truck, fmt.Sprintf("REC %s", truck.TerminalId))
}

func ForwardToClassification(entity sim.Entity) {
    env := entity.GetEnvironment()
    truck := sim.Cast[*Truck](entity)
    env.ForwardTo(truck, fmt.Sprintf("CLA %s", truck.TerminalId))
}

type TruckSource struct {
    sim.EntitySourceBase
    TerminalId      string
    CommodityId     string
    Month           int
}

type ShipSource struct {
    sim.EntitySourceBase
    HarborId        string
    CommodityId     string
    Month           int
}

func (source *TruckSource) Generate() sim.Entity {
    env := source.GetEnvironment()
    terminal := g.Terminals[source.TerminalId]
    
    truck := &Truck{
        TerminalId: source.TerminalId,
        CommodityId: source.CommodityId,
        Load: g.TruckCapacity,
    }
    
    terminal.Exports[source.CommodityId][source.Month] += truck.Load
    
    env.AddEntity("Truck", truck)
    env.ForwardTo(truck, fmt.Sprintf("ARR %s", source.TerminalId))
    
    return truck
}

func (source *ShipSource) Generate() sim.Entity {
    env := source.GetEnvironment()
    harbor := g.Harbors[source.HarborId]
    
    ship := &Ship{
        HarborId: source.HarborId,
        CommodityId: source.CommodityId,
        DWT: harbor.Productivity[source.CommodityId].DWT,
    }
    
    DWTOptions := []float64{50*KTon, 60*KTon, 70*KTon}
    harbor.Exports[source.CommodityId][source.Month] += DWTOptions[int(harbor.ShipDWTRNG.Next())]
    
    env.AddEntity("Ship", ship)
    env.ForwardTo(ship, fmt.Sprintf("DCK %s %s", source.CommodityId, source.HarborId))
    
    return ship
}

type Global struct {
    Env             *sim.Environment
    Commodities     map[string]*Commodity
    Terminals       map[string]*Terminal
    Harbors         map[string]*Harbor
    RailSections    map[string]map[string]float64
    TruckCapacity   float64
}

var g Global

func Check(err error) {
    if err != nil {
        log.Fatalf("ERROR: %s", err)
    }
}

func ReadTerminalExports(commId string, filePath string) {
    file, err := os.Open(filePath)
    Check(err)
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comma = '\t'
    records, err := reader.ReadAll()
    Check(err)
    
    g.Commodities[commId] = &Commodity{}
    g.Commodities[commId].Id = commId
    
    for _, row := range records[1:len(records)-1] {
        tid := row[0]
        terminal := g.Terminals[tid]
        terminal.MonthExports[commId] = &[12]float64{}
        terminal.Sazonality[commId] = &[12]float64{}
        terminal.NumTrucks[commId] = &[12]int{}
        terminal.Exports[commId] = &[12]float64{}
        terminal.AnnualExports[commId] = 0
        
        for mon, col := range row[1:len(row)-1] {
            value, err := strconv.ParseFloat(col, 64)
            terminal.Sazonality[commId][mon] = 0.0
            terminal.MonthExports[commId][mon] = 0.0
            if err == nil {
                terminal.MonthExports[commId][mon] = value*KTon
                terminal.AnnualExports[commId] += value*KTon
            }
        }
    }
    
    for m := 0; m < 12; m++ {
        for _, terminal := range g.Terminals {
            for cid, _ := range terminal.MonthExports {
                terminal.Sazonality[cid][m] = terminal.MonthExports[cid][m] / terminal.AnnualExports[cid]
                terminal.NumTrucks[cid][m] = int(math.Ceil((terminal.AnnualExports[cid] * terminal.Sazonality[cid][m]) / g.TruckCapacity))
            }
        }
    }
}

func ReadHarborExports(cid string, filePath string) {
    file, err := os.Open(filePath)
    Check(err)
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comma = '\t'
    records, err := reader.ReadAll()
    Check(err)
    
    for _, row := range records[1:len(records)-1] {
        hid := row[0]
        harbor := g.Harbors[hid]
        harbor.MonthExports[cid] = &[12]float64{}
        harbor.Sazonality[cid] = &[12]float64{}
        harbor.NumShips[cid] = &[12]int{}
        harbor.AnnualExports[cid] = 0
        harbor.Exports[cid] = &[12]float64{}
        
        for mon, col := range row[1:len(row)-1] {
            value, err := strconv.ParseFloat(col, 64)
            harbor.Sazonality[cid][mon] = 0.0
            harbor.MonthExports[cid][mon] = 0.0
            if err == nil {
                harbor.MonthExports[cid][mon] = value*KTon
                harbor.AnnualExports[cid] += value*KTon
            }
        }
    }
    
    for m := 0; m < 12; m++ {
        for _, harbor := range g.Harbors {
            for cid, _ := range harbor.MonthExports {
                harbor.Sazonality[cid][m] = harbor.MonthExports[cid][m] / harbor.AnnualExports[cid]
                harbor.NumShips[cid][m] = int(math.Ceil((harbor.AnnualExports[cid] * harbor.Sazonality[cid][m]) / harbor.Productivity[cid].DWT))
            }
        }
    }
}

func ReadHarborCommodityProductivity(cid string, filePath string) {
    file, err := os.Open(filePath)
    Check(err)
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comma = '\t'
    records, err := reader.ReadAll()
    Check(err)
    
    for _, row := range records[1:len(records)] {
        hid := row[0]
        
        harbor := g.Harbors[hid]
        prod := &CommodityProductivityInHarbor{}
        prod.DWT, _ = strconv.ParseFloat(row[1], 64)
        prod.OperatingTime, _ = strconv.ParseFloat(row[2], 64)
        prod.IdleTime, _ = strconv.ParseFloat(row[3], 64)
        prod.DockingTime, _ = strconv.ParseFloat(row[4], 64)
        prod.Productivity, _ = strconv.ParseFloat(row[5], 64)
        
        prod.DWT *= KTon
        prod.OperatingTime *= Hours
        prod.IdleTime *= Hours
        prod.DockingTime *= Hours
        prod.Productivity /= Hours
        
        harbor.Productivity[cid] = prod
        
        resources := math.Ceil((prod.DWT / prod.Productivity) / Days)+2
        
        g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("DCK %s", hid), Amount: resources})
        g.Env.AddProcess(
            sim.ProcessBase{
                Id: fmt.Sprintf("DCK %s %s", cid, hid),
                Needs: map[string]float64{
                    fmt.Sprintf("DCK %s", hid): 1,
                },
                RNG: sim.NewRNGTriangular(prod.DockingTime*0.95, prod.DockingTime*1.05, prod.DockingTime),
            },
        )
    }
}

func ReadData() {
    // Load terminal data
    //---------------------
    file, err := os.Open("data/terminals.tsv")
    Check(err)
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comma = '\t'
    records, err := reader.ReadAll()
    Check(err)
    
    for _, row := range records[1:len(records)] {
        tid := row[0]
        
        terminal := &Terminal{
            Id: tid,
            MonthExports: make(map[string]*[12]float64),
            Sazonality: make(map[string]*[12]float64),
            NumTrucks: make(map[string]*[12]int),
            Exports: make(map[string]*[12]float64),
            AnnualExports: make(map[string]float64),
        }
        
        terminal.Storage, _ = strconv.ParseFloat(row[1], 64)
        terminal.ProcessingRate, _ = strconv.ParseFloat(row[2], 64)
        
        terminal.Storage *= KTon
        terminal.ProcessingRate *= KTon
        
        trucksPerMinuteCap := math.Ceil(terminal.ProcessingRate / 30 / 30 / 24 / 60)
        resources := trucksPerMinuteCap+1
        
        g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("ARR %s", tid), Amount: resources})
        g.Env.AddProcess(
            sim.ProcessBase{
                Id: fmt.Sprintf("ARR %s", tid),
                Needs: map[string]float64{
                    fmt.Sprintf("ARR %s", tid): 1,
                },
                RNG: sim.NewRNGNormal(1.2*Minutes, 0.1*Minutes),
                NextProcess: fmt.Sprintf("REC %s", tid),
            },
        )
        
        g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("REC %s", tid), Amount: resources})
        g.Env.AddProcess(
            sim.ProcessBase{
                Id: fmt.Sprintf("REC %s", tid),
                Needs: map[string]float64{
                    fmt.Sprintf("REC %s", tid): 1,
                },
                RNG: sim.NewRNGLogNormal(1.9*Minutes, 0.7*Minutes),
                NextProcess: fmt.Sprintf("CLA %s", tid),
            },
        )
        
        g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("CLA %s", tid), Amount: resources})
        g.Env.AddProcess(
            sim.ProcessBase{
                Id: fmt.Sprintf("CLA %s", tid),
                Needs: map[string]float64{
                    fmt.Sprintf("CLA %s", tid): 1,
                },
                RNG: sim.NewRNGTriangular(1*Minutes, 5*Minutes, 3*Minutes),
                NextProcess: fmt.Sprintf("UNL %s", tid),
            },
        )
        
        g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("UNL %s", tid), Amount: resources})
        g.Env.AddProcess(
            sim.ProcessBase{
                Id: fmt.Sprintf("UNL %s", tid),
                Needs: map[string]float64{
                    fmt.Sprintf("UNL %s", tid): 1,
                },
                RNG: sim.NewRNGLogNormal(4.3*Minutes, 0.6*Minutes),
                NextProcess: fmt.Sprintf("EXI %s", tid),
            },
        )
        
        g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("EXI %s", tid), Amount: resources})
        g.Env.AddProcess(
            sim.ProcessBase{
                Id: fmt.Sprintf("EXI %s", tid),
                Needs: map[string]float64{
                    fmt.Sprintf("EXI %s", tid): 1,
                },
                RNG: sim.NewRNGNormal(1.2*Minutes, 0.1*Minutes),
            },
        )
        
        g.Terminals[tid] = terminal
    }
    
    // Load harbor data
    //---------------------
    file, err = os.Open("data/harbors.tsv")
    Check(err)
    defer file.Close()

    reader = csv.NewReader(file)
    reader.Comma = '\t'
    records, err = reader.ReadAll()
    Check(err)
    
    for _, row := range records[1:len(records)] {
        hid := row[0]
        
        harbor := &Harbor{
            Id: hid,
            MonthExports: make(map[string]*[12]float64),
            Sazonality: make(map[string]*[12]float64),
            NumShips: make(map[string]*[12]int),
            AnnualExports: make(map[string]float64),
            Exports: make(map[string]*[12]float64),
            Productivity: make(map[string]*CommodityProductivityInHarbor),
            ShipDWTRNG: *sim.NewRNGDiscrete([]float64{0.1, 0.55, 0.35}),
        }
        
        harbor.Storage, _ = strconv.ParseFloat(row[1], 64)
        harbor.Docks, _ = strconv.ParseFloat(row[2], 64)
        harbor.DockingInterval, _ = strconv.ParseFloat(row[3], 64)
        harbor.ProcessingRate, _ = strconv.ParseFloat(row[4], 64)
        
        harbor.Storage *= KTon
        harbor.DockingInterval *= Hours
        harbor.ProcessingRate *= KTon
        
        g.Harbors[hid] = harbor
    }
    
    // Load railway segments
    //--------------------------
    file, err = os.Open("data/segmentos-RMS.tsv")
    Check(err)
    defer file.Close()

    reader = csv.NewReader(file)
    reader.Comma = '\t'
    records, err = reader.ReadAll()
    Check(err)
    
    for _, row := range records[1:len(records)] {
        a, ok := g.RailSections[row[0]]
        if !ok {
            a = make(map[string]float64)
            g.RailSections[row[0]] = a
        }
        
        a[row[1]], _ = strconv.ParseFloat(row[2], 64)
        
        b, ok := g.RailSections[row[1]]
        if !ok {
            b = make(map[string]float64)
            g.RailSections[row[1]] = b
        }
        
        b[row[0]], _ = strconv.ParseFloat(row[2], 64)
    }
    
    //pretty.Println(g.RailSections)
    
    for id1, _ := range g.RailSections {
        for id2, _ := range g.RailSections {
            g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("RSC %s %s", id1, id2), Amount: 1})
            g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("RSC %s %s", id2, id1), Amount: 1})
        
            g.Env.AddProcess(
                sim.ProcessBase{
                    Id: fmt.Sprintf("TVL %s %s", id1, id2),
                    Group: "TVL",
                    Needs: map[string]float64{
                        fmt.Sprintf("RSC %s %s", id1, id2): 1,
                        fmt.Sprintf("RSC %s %s", id2, id1): 1,
                    },
                    RNG: sim.NewRNGTriangular(1, 5, 3),
                },
            )
            
            g.Env.AddProcess(
                sim.ProcessBase{
                    Id: fmt.Sprintf("TVL %s %s", id2, id1),
                    Group: "TVL",
                    Needs: map[string]float64{
                        fmt.Sprintf("RSC %s %s", id1, id2): 1,
                        fmt.Sprintf("RSC %s %s", id2, id1): 1,
                    },
                    RNG: sim.NewRNGTriangular(1, 5, 3),
                },
            )
        }
    }
    
    train := &Train{
        TerminalId: "",
        CommodityId: "",
        Load: g.TruckCapacity,
    }
    
    _ = train
    
    ReadHarborCommodityProductivity("Corn", "data/harbor-corn-cap.tsv")
    ReadHarborCommodityProductivity("Soy", "data/harbor-soy-cap.tsv")
    
    ReadTerminalExports("Corn", "data/corn.tsv")
    ReadTerminalExports("Soy", "data/soy.tsv")
    
    ReadHarborExports("Corn", "data/harbor-corn-exports-2022.tsv")
    ReadHarborExports("Soy", "data/harbor-soy-exports-2022.tsv")
    
    //pretty.Println(g.Harbors)
}

func PrintExports() {
    // Print terminal exports
    for cid, _ := range g.Commodities {
        fmt.Printf("[TERMINAL EXPORTS] (%s kt)\n", cid)
        
        fmt.Printf("%24s", "Terminal")
        
        for m := 0; m < 12; m++ {
            fmt.Printf("%9s", MonthName[m])
        }
        
        fmt.Println()
        
        for tid, terminal := range g.Terminals {
            fmt.Printf("%24s", tid)
            
            for m := 0; m < 12; m++ {
                fmt.Printf("%9.2f", terminal.Exports[cid][m] / KTon)
            }
            
            fmt.Println()
        }
        
        fmt.Println()
    }
    
    // Print harbor exports
    for cid, _ := range g.Commodities {
        fmt.Printf("[HARBOR EXPORTS] (%s kt)\n", cid)
        
        fmt.Printf("%24s", "Harbor")
        
        for m := 0; m < 12; m++ {
            fmt.Printf("%9s", MonthName[m])
        }
        
        fmt.Println()
        
        for hid, harbor := range g.Harbors {
            fmt.Printf("%24s", hid)
            
            for m := 0; m < 12; m++ {
                fmt.Printf("%9.2f", harbor.Exports[cid][m] / KTon)
            }
            
            fmt.Println()
        }
        
        fmt.Println()
    }
}

func main() {
    g.Commodities = make(map[string]*Commodity)
    g.Terminals = make(map[string]*Terminal)
    g.Harbors = make(map[string]*Harbor)
    g.RailSections = make(map[string]map[string]float64)
    g.TruckCapacity = 30
    
    g.Env = sim.NewEnvironment()
    
    ReadData()
    
    g.Env.Now = 0.0
    
    day := 0
    for m := 0; m < 12; m++ {
        for tid, terminal := range g.Terminals {
            for cid, _ := range g.Commodities {
                if terminal.MonthExports[cid][m] > 0.0 {
                    interval := float64(NumDays[m])*Days / float64(terminal.NumTrucks[cid][m])
                    sid := fmt.Sprintf("%s:%s:%d", tid, cid, m+1)
                    
                    g.Env.AddEntitySource(&TruckSource{
                        EntitySourceBase: sim.EntitySourceBase{
                            Id: sid, 
                            NextGen: float64(day)*Days,
                            RNG: sim.NewRNGExponential(1/interval),
                            BatchSize: 1, 
                            MaxGenerations: terminal.NumTrucks[cid][m],
                        },
                        TerminalId: tid,
                        CommodityId: cid,
                        Month: m,
                    })
                }
                
            }
        }
        
        for hid, harbor := range g.Harbors {
            for cid, _ := range g.Commodities {
                if harbor.MonthExports[cid][m] > 0.0 {
                    interval := float64(NumDays[m])*Days / float64(harbor.NumShips[cid][m])
                    sid := fmt.Sprintf("%s:%s:%d", hid, cid, m+1)
                    
                    g.Env.AddEntitySource(&ShipSource{
                        EntitySourceBase: sim.EntitySourceBase{
                            Id: sid, 
                            NextGen: float64(day)*Days,
                            RNG: sim.NewRNGExponential(1/interval),
                            BatchSize: 1,
                            MaxGenerations: harbor.NumShips[cid][m],
                        },
                        HarborId: hid,
                        CommodityId: cid,
                        Month: m,
                    })
                }
            }
        }
        
        day += NumDays[m]
    }
    
    g.Env.LogLevel = 1
    g.Env.StepThrough = false
    g.Env.EndDate = 30*Days
    
    g.Env.Run()
    //g.Env.PrintProcessStatistics()
    g.Env.PrintProcessGroupStatistics("")
    
    fmt.Println()
    PrintExports()
}










