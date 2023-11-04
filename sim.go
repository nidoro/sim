package sim

import (
    "fmt"
    "slices"
    "time"
    "math"
    "sort"
    "bufio"
    "os"
    "gonum.org/v1/gonum/stat/distuv"
    "golang.org/x/exp/rand"
)

const (
    AC_Reset         string = "\x1b[0m"
    AC_CodeRed       string = "\x1b[31m"
    AC_CodeGreen     string = "\x1b[32m"
    AC_CodeYellow    string = "\x1b[33m"
    AC_CodeBlue      string = "\x1b[34m"
    AC_CodeMagenta   string = "\x1b[35m"
    AC_CodeCyan      string = "\x1b[36m"
    AC_CodeBold      string = "\x1b[1m"
    AC_ResetBold     string = "\x1b[22m"
)

func AC_Red(str string) string {return AC_CodeRed + str + AC_Reset}
func AC_Green(str string) string {return AC_CodeGreen + str + AC_Reset}
func AC_Yellow(str string) string {return AC_CodeYellow + str + AC_Reset}
func AC_Blue(str string) string {return AC_CodeBlue + str + AC_Reset}
func AC_Magenta(str string) string {return AC_CodeMagenta + str + AC_Reset}
func AC_Cyan(str string) string {return AC_CodeCyan + str + AC_Reset}
func AC_Bold(str string) string {return AC_CodeBold + str + AC_ResetBold}

func WaitForEnter() {
    fmt.Printf(AC_Bold("[STEP THROUGH] Press ENTER to continue\n"))
    bufio.NewReader(os.Stdin).ReadBytes('\n')
}

type RNG interface {
    Next() float64
}

type RNGExponential struct {
    Rate    float64
    RNG     rand.Rand
}

type RNGNormal struct {
    Mean    float64
    StdDev  float64
    RNG     rand.Rand
}

type RNGLogNormal struct {
    Mean    float64
    StdDev  float64
    RNG     distuv.LogNormal
}

type RNGTriangular struct {
    A       float64
    B       float64
    C       float64
    RNG     distuv.Triangle
}

type RNGDiscrete struct {
    Weights []float64
    RNG     distuv.Categorical
}

func NewRNGExponential(rate float64) *RNGExponential {
    return &RNGExponential{Rate: rate, RNG: *rand.New(rand.NewSource(uint64(time.Now().UnixMilli())))}
}

func NewRNGNormal(mean float64, stddev float64) *RNGNormal {
    return &RNGNormal{Mean: mean, StdDev: stddev, RNG: *rand.New(rand.NewSource(uint64(time.Now().UnixMilli())))}
}

func NewRNGLogNormal(mean float64, stddev float64) *RNGLogNormal {
    mu := math.Log(math.Pow(mean, 2) / math.Sqrt(math.Pow(mean, 2) + math.Pow(stddev, 2)))
    sigma := math.Sqrt(math.Log(1 + math.Pow(stddev, 2)/math.Pow(mean, 2)))
    return &RNGLogNormal{Mean: mean, StdDev: stddev, RNG: distuv.LogNormal{Mu: mu, Sigma: sigma, Src: rand.NewSource(uint64(time.Now().UnixMilli()))}}
}

func NewRNGTriangular(a float64, b float64, c float64) *RNGTriangular {
    return &RNGTriangular{A: a, B: b, C: c, RNG: distuv.NewTriangle(a, b, c, rand.NewSource(uint64(time.Now().UnixMilli())))}
}

func NewRNGDiscrete(w []float64) *RNGDiscrete {
    return &RNGDiscrete{Weights: w, RNG: distuv.NewCategorical(w, rand.NewSource(uint64(time.Now().UnixMilli())))}
}

func (rng *RNGExponential) Next() float64 {
    return rng.RNG.ExpFloat64() / rng.Rate
}

func (rng *RNGNormal) Next() float64 {
    return rng.RNG.NormFloat64() * rng.StdDev + rng.Mean
}

func (rng *RNGLogNormal) Next() float64 {
    return rng.RNG.Rand()
}

func (rng *RNGTriangular) Next() float64 {
    return rng.RNG.Rand()
}

func (rng *RNGDiscrete) Next() float64 {
    return rng.RNG.Rand()
}

type QueueType int

const (
    QueueType_Resource QueueType = iota
    QueueType_Process
)

type QueueStatistics struct {
    TotalEntitiesIn int
    TotalEntitiesOut int
    TotalTimeInQueue float64
    AvgTimeInQueue float64
}

type QueueStats struct {
    Type        QueueType
    Id          string
    DateIn      float64
    DateOut     float64
}

type ProcessStats struct {
    Id              string
    DateQueued      float64
    DateStart       float64
    DateEnd         float64
}

type EntityBase struct {
    Id      int
    Type    string
    QueueStats   []*QueueStats
    ProcessStats []*ProcessStats
    Resources    map[string]float64
    Environment  *Environment
}

type Entity interface {
    GetEnvironment() *Environment
    GetEntityBase() *EntityBase
    Initialize(id int, tp string)
    SetId(id int)
    GetId() int
    SetType(tp string)
    GetType() string
    GetName() string
    
    EnterQueue(queueType QueueType, id string, date float64)
    LeaveQueue(queueType QueueType, id string, date float64)
    StartProcess(date float64)
    EndProcess(date float64)
    GetTimeInQueue() float64
    
    GetResourceAmount(rid string) float64
    SeizeResource(rid string, amount float64, date float64)
    ReleaseResources()
}

func (entityBase *EntityBase) GetEnvironment() *Environment {
    return entityBase.Environment
}

func (entityBase *EntityBase) GetEntityBase() *EntityBase {
    return entityBase
}

func (entityBase *EntityBase) GetResourceAmount(rid string) float64 {
    amount, ok := entityBase.Resources[rid]
    if ok {
        return amount
    }
    return 0
}

func (entityBase *EntityBase) GetTimeInQueue() float64 {
    st := entityBase.ProcessStats[len(entityBase.ProcessStats)-1]
    return st.DateStart - st.DateQueued
}

func (entityBase *EntityBase) StartProcess(date float64) {
    entityBase.ProcessStats[len(entityBase.ProcessStats)-1].DateStart = date
}

func (entityBase *EntityBase) EndProcess(date float64) {
    entityBase.ProcessStats[len(entityBase.ProcessStats)-1].DateEnd = date
}

func (entityBase *EntityBase) Initialize(id int, tp string) {
    entityBase.Id = id
    entityBase.Type = tp
    entityBase.QueueStats = make([]*QueueStats, 0)
    entityBase.ProcessStats = make([]*ProcessStats, 0)
    entityBase.Resources = make(map[string]float64)
}

func (entityBase *EntityBase) SetId(id int) {
    entityBase.Id = id
}

func (entityBase *EntityBase) GetId() int {
    return entityBase.Id
}

func (entityBase *EntityBase) SetType(tp string) {
    entityBase.Type = tp
}

func (entityBase *EntityBase) GetType() string {
    return entityBase.Type
}

func (entityBase *EntityBase) GetName() string {
    return fmt.Sprintf("%s %d", entityBase.Type, entityBase.Id)
}

func (entityBase *EntityBase) EnterQueue(tp QueueType, id string, date float64) {
    entityBase.QueueStats = append(entityBase.QueueStats, &QueueStats{Type: tp, Id: id, DateIn: date})
    if tp == QueueType_Process {
        entityBase.ProcessStats = append(entityBase.ProcessStats, &ProcessStats{Id: id, DateQueued: date})
    }
}

func (entityBase *EntityBase) LeaveQueue(tp QueueType, id string, date float64) {
    for i := len(entityBase.QueueStats)-1; i >= 0; i-- {
        st := entityBase.QueueStats[i]
        if st.Type == tp && st.Id == id {
            st.DateOut = date
            return
        }
    }
}

func (entityBase *EntityBase) SeizeResource(rid string, amount float64, date float64) {
    entityBase.Resources[rid] = amount
    entityBase.LeaveQueue(QueueType_Resource, rid, date)
}

func (entityBase *EntityBase) ReleaseResources() {
    entityBase.Resources = make(map[string]float64)
}

type ResourceBase struct {
    Id          string
    Amount      float64
    Queue       []Entity
    
    // Statistics
    TotalEntitiesIn int
    TotalEntitiesOut int
    TotalTimeInQueue float64
    AvgTimeInQueue float64
}

type Resource interface {
    Enqueue(entity Entity)
    Dequeue()
    GetAmount() float64
    SetAmount(amount float64)
}

func (res *ResourceBase) Enqueue(entity Entity) {
    res.Queue = append(res.Queue, entity)
    res.TotalEntitiesIn++
}

func (res *ResourceBase) Dequeue() {
    res.Queue = res.Queue[1:]
    res.TotalEntitiesOut++
}

func (res *ResourceBase) GetAmount() float64 {
    return res.Amount
}

func (res *ResourceBase) SetAmount(amount float64) {
    res.Amount = amount
}

type ProcessBase struct {
    Id          string
    Group       string
    Needs       map[string]float64
    Queue       []Entity
    RNG         RNG
    QueueStats  QueueStatistics
    Forward     func (entity Entity)
    NextProcess string
}

type Process interface {
    GetProcessBase() *ProcessBase
    Initialize(base ProcessBase)
    GetId() string
    GetDuration(entity Entity) float64
    GetNeeds() map[string]float64
    Enqueue(entity Entity)
    Dequeue()
    GetQueueSize() int
    GetNextInQueue() Entity
    GetStatistics() QueueStatistics
}

func (process *ProcessBase) GetProcessBase() *ProcessBase {
    return process
}

func (process *ProcessBase) Initialize(base ProcessBase) {
    *process = base
}

func (process *ProcessBase) GetStatistics() QueueStatistics {
    return process.QueueStats
}

func (process *ProcessBase) GetId() string {
    return process.Id
}

func (process *ProcessBase) GetNeeds() map[string]float64 {
    return process.Needs
}

func (process *ProcessBase) GetDuration(entity Entity) float64 {
    return process.RNG.Next()
}

func (process *ProcessBase) Enqueue(entity Entity) {
    process.Queue = append(process.Queue, entity)
    process.QueueStats.TotalEntitiesIn++
}

func (process *ProcessBase) Dequeue() {
    entity := process.Queue[0]
    process.QueueStats.TotalTimeInQueue += entity.GetTimeInQueue()
    process.Queue = process.Queue[1:]
    process.QueueStats.TotalEntitiesOut++
    process.QueueStats.AvgTimeInQueue = process.QueueStats.TotalTimeInQueue / float64(process.QueueStats.TotalEntitiesIn)
}

func (process *ProcessBase) GetQueueSize() int {
    return len(process.Queue)
}

func (process *ProcessBase) GetNextInQueue() Entity {
    return process.Queue[0]
}
type OngoingProcess struct {
    Process Process
    Entity Entity
    DateStart float64
    DateEnd float64
}

type ByDateEnd []OngoingProcess

func (a ByDateEnd) Len() int           { return len(a) }
func (a ByDateEnd) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDateEnd) Less(i, j int) bool { return a[i].DateEnd < a[j].DateEnd }

type EntitySourceBase struct {
    Id              string
    RNG             RNG
    MaxGenerations  int
    BatchSize       int
    Env             *Environment
    
    // Simulation
    NextGen         float64
    Generations     int
}

type EntitySource interface {
    GetEntitySourceBase() *EntitySourceBase
    GetEnvironment() *Environment
    Initialize(base EntitySourceBase)
    GetId()         string
    Generate()      Entity
    Update()
    GetNextGen()    float64
    GetBatchSize()  int
    GetMaxGenerations()  int
    GetGenerations()  int
}

type ByNextGen []EntitySource

func (a ByNextGen) Len() int           { return len(a) }
func (a ByNextGen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNextGen) Less(i, j int) bool { return a[i].GetNextGen() < a[j].GetNextGen() }

func (source *EntitySourceBase) GetEntitySourceBase() *EntitySourceBase {
    return source
}

func (source *EntitySourceBase) GetEnvironment() *Environment {
    return source.Env
}

func (source *EntitySourceBase) GetId() string {
    return source.Id
}

func (source *EntitySourceBase) Initialize(base EntitySourceBase) {
    *source = base
}

func (source *EntitySourceBase) GetNextGen() float64 {
    return source.NextGen
}

func (source *EntitySourceBase) GetBatchSize() int {
    return source.BatchSize
}

func (source *EntitySourceBase) GetMaxGenerations() int {
    return source.MaxGenerations
}

func (source *EntitySourceBase) GetGenerations() int {
    return source.Generations
}

func (source *EntitySourceBase) Update() {
    source.NextGen += source.RNG.Next()
    source.Generations++
}

type PrintfFunc func (format string, a ...any) (n int, err error)
func DisabledPrintf(format string, a ...any) (n int, err error) {return 0, nil}

type Environment struct {
    EntitySources   []EntitySource // array because needs sorting
    Resources       map[string]Resource // map of strings because persistent
    Entities        map[int]Entity // map of int because constantly deleting
    Processes       []Process // array because needs sorting
    WatchedProcesses map[string]Process // array because needs sorting
    OngoingProcesses []OngoingProcess // array because needs sorting
    NextEntityId    int
    Now             float64 // seconds
    EndDate         float64 // seconds
    
    StepThrough      bool
    LogLevel        int
    Printf          [3]PrintfFunc
}

func (env *Environment) SetLogLevel(level int) {
    for i := 0; i < len(env.Printf); i++ {
        env.Printf[i] = DisabledPrintf
    }
    
    for i := 1; i <= level; i++ {
        env.Printf[i] = fmt.Printf
    }
    
    env.LogLevel = level
}

func GetHumanTime(s float64) string {
    days := s / 60 / 60 / 24
    hours := (days - math.Floor(days)) * 24
    minutes := (hours - math.Floor(hours)) * 60
    seconds := (minutes - math.Floor(minutes)) * 60
    
    return fmt.Sprintf("%.0fd %02.0f:%02.0f:%05.2f", math.Floor(days), math.Floor(hours), math.Floor(minutes), seconds)
}

func (env *Environment) Enqueue(entity Entity, process Process) {
    for rid, _ := range process.GetNeeds() {
        env.Resources[rid].Enqueue(entity)
        entity.EnterQueue(QueueType_Resource, rid, env.Now)
    }
    
    process.Enqueue(entity)
    entity.EnterQueue(QueueType_Process, process.GetId(), env.Now)
    
    env.WatchedProcesses[process.GetId()] = process
}

func (env *Environment) GetProcess(pid string) Process {
    for _, process := range env.Processes {
        if process.GetId() == pid {
            return process
        }
    }
    return nil
}

func (env *Environment) ForwardTo(entity Entity, pid string) {
    env.Enqueue(entity, env.GetProcess(pid))
}

func (env *Environment) AddResource(resource *ResourceBase) {
    env.Resources[resource.Id] = resource
}

func (env *Environment) AddProcess(base ProcessBase) {
    env.Processes = append(env.Processes, &base)
}
    
func (env *Environment) AddEntitySource(entitySource EntitySource) {
    entitySource.GetEntitySourceBase().Env = env
    env.EntitySources = append(env.EntitySources, entitySource)
}

func (env *Environment) AddEntity(entityType string, entity Entity) {
    entity.Initialize(env.NextEntityId, entityType)
    env.Entities[entity.GetId()] = entity
    env.NextEntityId++
}
func (env *Environment) MaybeStartProcess(process Process) {
    for process.GetQueueSize() > 0 {
        entity := process.GetNextInQueue()
        readyToStart := true
        
        for rid, amount := range process.GetNeeds() {
            seized := entity.GetResourceAmount(rid)
            if seized < amount {
                if env.Resources[rid].GetAmount() >= amount {
                    env.Resources[rid].SetAmount(env.Resources[rid].GetAmount() - amount)
                    entity.SeizeResource(rid, amount, env.Now)
                } else {
                    readyToStart = false
                }
            }
        }
        
        if readyToStart {
            env.Printf[2]("[PROCESS STARTED] %s | %s\n", process.GetId(), entity.GetName())
            entity.LeaveQueue(QueueType_Process, process.GetId(), env.Now)
            env.StartProcess(process, entity, env.Now + process.GetDuration(entity))
            
            if process.GetQueueSize() == 0 {
                env.WatchedProcesses[process.GetId()] = nil
            }
        } else {
            break
        }
    }
}

func (env *Environment) StartProcess(process Process, entity Entity, endDate float64) {
    entity.StartProcess(env.Now)
    process.Dequeue()
    ongoing := OngoingProcess{Process: process, Entity: entity, DateStart: env.Now, DateEnd: endDate}
    env.OngoingProcesses = append(env.OngoingProcesses, ongoing)
}

func Cast[T Entity](entity Entity) T {
    if t, ok := entity.(T); ok {
        return t
    }
    panic("Cast failed!")
}

func (env *Environment) Run() {
    if env.StepThrough {
        env.LogLevel = 2
    }
    
    env.SetLogLevel(env.LogLevel)
    sort.Sort(ByNextGen(env.EntitySources))
    env.Now = env.EntitySources[0].GetNextGen()
    
    runStart := time.Now()
    lastBarRefresh := time.Now()
    
    env.Printf[1]("[STARTING SIMULATION]\n")
    env.Printf[1]("[MAX TIME] %s\n", GetHumanTime(env.EndDate))
    
    for env.Now < env.EndDate {
        env.Printf[2](AC_Green(AC_Bold("[SIMULATION CLOCK] %s (%.2fs)\n")), GetHumanTime(env.Now), env.Now)
        
        for s := 0; s < len(env.EntitySources); {
            source := env.EntitySources[s]
            
            if source.GetNextGen() > env.Now {
                break
            }
            
            for e := 0; e < source.GetBatchSize(); e++ {
                entity := source.Generate()
                env.Printf[2]("[NEW ENTITY] %s | %s\n", source.GetId(), entity.GetName())
                _ = entity
            }
            
            source.Update()
            
            if source.GetGenerations() == source.GetMaxGenerations() {
                env.EntitySources = slices.Delete(env.EntitySources, s, s+1)
            } else {
                s++
            }
        }
        
        nextTime := env.Now
        
        for nextTime == env.Now {
            for len(env.OngoingProcesses) > 0 {
                ongoing := env.OngoingProcesses[0]
                
                if ongoing.DateEnd > env.Now {
                    break
                }
                
                entity := ongoing.Entity
                process := ongoing.Process
                
                entity.EndProcess(env.Now)
                env.Printf[2]("[PROCESS ENDED] %s | %s\n", process.GetId(), entity.GetName())
                
                for rid, amount := range process.GetNeeds() {
                    env.Resources[rid].SetAmount(env.Resources[rid].GetAmount() + amount)
                }
                
                entity.ReleaseResources()
                env.OngoingProcesses = env.OngoingProcesses[1:]
                
                if ongoing.Process.GetProcessBase().Forward != nil {
                    ongoing.Process.GetProcessBase().Forward(entity)
                } else if ongoing.Process.GetProcessBase().NextProcess != "" {
                    env.ForwardTo(entity, ongoing.Process.GetProcessBase().NextProcess)
                } else {
                    // dispose
                }
            }
            
            // start processes that can be started
            for _, process := range env.WatchedProcesses {
                env.MaybeStartProcess(process)
            }
            
            for key, process := range env.WatchedProcesses {
                if process == nil {
                    delete(env.WatchedProcesses, key)
                }
            }
            
            //fmt.Println(env.WatchedProcesses)
            
            sort.Sort(ByDateEnd(env.OngoingProcesses))
            sort.Sort(ByNextGen(env.EntitySources))
            
            // Update simulation clock
            nextTime = env.EndDate
            if len(env.EntitySources) > 0 {
                nextTime = min(nextTime, env.EntitySources[0].GetNextGen())
            }
            
            if len(env.OngoingProcesses) > 0 {
                nextTime = min(nextTime, env.OngoingProcesses[0].DateEnd)
            }
        }
        
        env.Now = nextTime
        
        if env.StepThrough {
            WaitForEnter()
        } else if (env.LogLevel == 1) {
            if time.Since(lastBarRefresh).Seconds() >= 1.0/15.0 {
                lastBarRefresh = time.Now()
                progress := env.Now / env.EndDate
                progressBarMax := 40
                
                fmt.Printf("\033[%dD[", progressBarMax+2)
                
                progressBarSize := int(math.Ceil(progress * float64(progressBarMax)))
                
                for c := 0; c < progressBarSize; c++ {
                    env.Printf[1]("\u25A0")
                }
                env.Printf[1]("%.*s]", progressBarMax - progressBarSize, "                                            ")
            }
        } else {
            env.Printf[2]("\n")
        }
    }
    
    env.Printf[1]("\n")
    env.Printf[1]("[SIMULATION ENDED]\n")
    env.Printf[1]("[RUN TIME] %.2fs\n", time.Since(runStart).Seconds())
    
    env.Printf[1]("\n")
}

func (env *Environment) PrintProcessGroupStatistics(groupId string) {
    fmt.Printf("[PROCESS STATISTICS] Group: %s\n", groupId)
    
    fmt.Printf("%24s%16s%16s%16s\n", "Process", "Entities In", "Entities Out", "Avg Q Time (s)")
        
    for _, process := range env.Processes {
        if process.GetProcessBase().Group == groupId {
            st := process.GetStatistics()
            fmt.Printf("%24.24s%16d%16d%16.2f\n", process.GetId(), st.TotalEntitiesIn, st.TotalEntitiesOut, st.AvgTimeInQueue)
        }
    }
}

func (env *Environment) PrintProcessStatistics() {
    fmt.Printf("[PROCESS STATISTICS]\n")
    fmt.Printf("%24s%16s%16s%16s\n", "Process", "Entities In", "Entities Out", "Avg Q Time (s)")
    
    for _, process := range env.Processes {
        st := process.GetStatistics()
        fmt.Printf("%24.24s%16d%16d%16.2f\n", process.GetId(), st.TotalEntitiesIn, st.TotalEntitiesOut, st.AvgTimeInQueue)
    }
}

func NewEnvironment() *Environment {
    env := &Environment{}
    env.Entities = make(map[int]Entity, 0)
    env.EntitySources = make([]EntitySource, 0)
    env.Resources = make(map[string]Resource, 0)
    env.Processes = make([]Process, 0)
    env.OngoingProcesses = make([]OngoingProcess, 0)
    env.WatchedProcesses = make(map[string]Process)
    return env
}

