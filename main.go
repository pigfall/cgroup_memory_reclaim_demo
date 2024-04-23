package main

import(
    "time"
    "os"
    "errors"
    "io"
    "os/exec"
    "strconv"
    "path/filepath"
    "fmt"
    "bytes"
    "bufio"
    "strings"
)

func main(){
  if isCgroupv2(){
    cgroupv2Demo()
  }else{
    cgroupv1Demo()
  }
}

func isCgroupv2()bool{
  output,err := exec.Command("mount").CombinedOutput()
  if err != nil{
    panic(err)
  }
  rd := bufio.NewReader(bytes.NewReader(output))
  for{
    line,err := rd.ReadBytes('\n')
    if err != nil{
      if errors.Is(err,io.EOF){
        return false
      }else{
        panic(err)
      }
    }
    elems := strings.Split(string(line)," ")
    if elems[2] == "/sys/fs/cgroup" && elems[4] == "cgroup2"{
      return true
    }
  }

}

func cgroupv2Demo(){
  demoCgroupPath := "/sys/fs/cgroup/demo"
  // Create a cgroup.
  if err := os.MkdirAll(demoCgroupPath,os.ModePerm);err != nil{
    panic(err)
  }
  // Move this pid to the created cgroup.
  pid := os.Getpid()
  if err := os.WriteFile(filepath.Join(demoCgroupPath,"cgroup.procs"),[]byte(fmt.Sprintf("%d",pid)),os.ModePerm);err != nil{
    panic(err)
  }
  // For incresing page/buffer cache , write 100M data to disk.
  if err := os.WriteFile("/tmp/demo_data",make([]byte,1024*1024*100),os.ModePerm);err != nil{
    panic(err)
  }
  defer os.Remove("/tmp/demo_data")
  // Inspect the current value of page/buffer cache
  fmt.Printf("Cache: %d\n", cacheInCgroupv2(filepath.Join(demoCgroupPath,"memory.stat")))
  // Trigger memory reclaim.
  // By setting the memory.high to a small value to trigger the memory reclaim. Do not forget to restore the value of memory.high, otherwise the cache will be frequently reclaim on the memory usage is higher than the setted value.
  restoreValue,err := os.ReadFile(filepath.Join(demoCgroupPath,"memory.high"))
  if err != nil{
    panic(err)
  }
  defer os.WriteFile(filepath.Join(demoCgroupPath,"memory.high",),restoreValue, os.ModePerm)
  if err := os.WriteFile(filepath.Join(demoCgroupPath,"memory.high",),[]byte("1024"), os.ModePerm);err != nil{
    panic(err)
  }
  time.Sleep(time.Second*2)
  fmt.Printf("Cache after reclaimed: %d\n", cacheInCgroupv2(filepath.Join(demoCgroupPath,"memory.stat")))
}

func cgroupv1Demo(){
  memorySubsystemPath := "/sys/fs/cgroup/memory/demo"
  // Create a memory subsystem.
  if err := os.MkdirAll(memorySubsystemPath,os.ModePerm);err !=nil{
    panic(err)
  }
  // Move this pid to the created memory subsystem.
  pid := os.Getpid()
  if err :=os.WriteFile(filepath.Join(memorySubsystemPath,"tasks"),[]byte(fmt.Sprintf("%d",pid)),os.ModePerm);err != nil{
    panic(err)
  }
  // For incresing page/buffer cache , write 100M data to disk.
  if err := os.WriteFile("/tmp/demo_data",make([]byte,1024*1024*100),os.ModePerm);err != nil{
    panic(err)
  }
  defer os.Remove("/tmp/demo_data")
  // Inspect the current value of page/buffer cache
  fmt.Printf("Cache: %d\n", cacheInCgroupv1(filepath.Join(memorySubsystemPath,"memory.stat")))
  // Trigger memory claim.
  if err := os.WriteFile(filepath.Join(memorySubsystemPath,"memory.force_empty",),[]byte("1"), os.ModePerm);err != nil{
    panic(err)
  }
  time.Sleep(time.Second*2)
  fmt.Printf("Cache after reclaimed: %d\n", cacheInCgroupv1(filepath.Join(memorySubsystemPath,"memory.stat")))
}

func cacheInCgroupv1(path string)int{
  content,err := os.ReadFile(path)
  if err != nil{
    panic(err)
  }
  rd := bufio.NewReader(bytes.NewReader(content))
  for{
    line,err := rd.ReadBytes('\n')
    if err != nil{
      panic(fmt.Errorf("Parse value of cache from file %s failed: %w",path,err))
    }
    lineStr := strings.TrimSuffix(string(line),"\n")
    elems := strings.SplitN(lineStr," ",2)
    if elems[0] == "cache"{
      value,err:=strconv.ParseInt(elems[1],10,64)
      if err != nil{
        panic(err)
      }
      return int(value)
    }
  }
}

func cacheInCgroupv2(path string)int{
  content,err := os.ReadFile(path)
  if err != nil{
    panic(err)
  }
  rd := bufio.NewReader(bytes.NewReader(content))
  for{
    line,err := rd.ReadBytes('\n')
    if err != nil{
      panic(fmt.Errorf("Parse value of cache from file %s failed: %w",path,err))
    }
    lineStr := strings.TrimSuffix(string(line),"\n")
    elems := strings.SplitN(lineStr," ",2)
    if elems[0] == "file"{
      value,err:=strconv.ParseInt(elems[1],10,64)
      if err != nil{
        panic(err)
      }
      return int(value)
    }
  }
}


