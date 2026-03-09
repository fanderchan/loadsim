# LoadSim

[![CI](https://github.com/fanderchan/loadsim/actions/workflows/ci.yml/badge.svg)](https://github.com/fanderchan/loadsim/actions/workflows/ci.yml)

`LoadSim` 是一个资源场景模拟工具，用于在 Linux 主机上构造可控的 CPU、内存，以及 CPU+内存组合占用场景。

正式产品名：

- 英文名：`LoadSim`
- 中文名：`负载场景模拟器`

命令行名称：

```bash
loadsim
```

## 这是什么

LoadSim 适合做下面几类事情：

- 固定占住一部分 CPU
- 固定占住一部分内存
- 同时占住 CPU 和内存
- 生成固定曲线
- 生成周期性波动曲线
- 运行时持续输出当前状态，方便观察

当前版本聚焦三条命令：

- `cpu`
- `ram`
- `combo`

## 功能概览

### `cpu`

制造 CPU 占用，支持两种模式：

- `fixed`：固定占用
- `wave`：在最小值和最大值之间周期性波动

### `ram`

制造内存占用，支持两种模式：

- `fixed`：固定占用
- `wave`：在最小值和最大值之间周期性波动

### `combo`

同时制造 CPU 和内存占用：

- CPU 可单独选择 `fixed` 或 `wave`
- RAM 可单独选择 `fixed` 或 `wave`
- 适合组合场景统一启动

## 编译

### 直接编译

```bash
go build -o loadsim .
```

### 使用脚本编译

```bash
./build.sh
```

构建完成后默认产物位置：

```bash
build/loadsim
```

## 首次使用

建议第一次先执行：

```bash
./loadsim --help
./loadsim cpu --help
./loadsim ram --help
./loadsim combo --help
./loadsim version
```

实际输出示例：

```text
$ ./loadsim --help
LoadSim is a resource occupancy CLI for creating controllable CPU, RAM, and combined resource scenarios.

Usage:
  loadsim [command]

Available Commands:
  combo       Occupy CPU and RAM at the same time
  completion  Generate the autocompletion script for the specified shell
  cpu         Occupy CPU with fixed or wave patterns
  help        Help about any command
  ram         Occupy RAM with fixed or wave patterns
  version     Print the current version
```

版本输出示例：

```text
$ ./loadsim version
LoadSim 0.2.0
```

## 快速开始

### 1. 固定占用 CPU

```bash
./loadsim cpu --mode fixed --percent 50 --cores 4
```

含义：

- 使用 4 个 worker
- 目标 CPU 占用约 50%
- 一直运行，直到手动停止

如果只运行 60 秒：

```bash
./loadsim cpu --mode fixed --percent 50 --cores 4 --time 60
```

### 2. 固定占用内存

```bash
./loadsim ram --mode fixed --size 1024
```

含义：

- 持续占用 1024MB 内存

### 3. 同时占用 CPU 和内存

```bash
./loadsim combo --cpu-percent 40 --cpu-cores 2 --ram-size 1024
```

含义：

- CPU 目标占用 40%
- CPU 使用 2 个 worker
- 内存占用 1024MB

### 4. 运行波动场景

CPU 波动：

```bash
./loadsim cpu --mode wave --min 20 --max 80 --period 60 --cores 4
```

RAM 波动：

```bash
./loadsim ram --mode wave --min-size 256 --max-size 2048 --period 120
```

CPU 和 RAM 同时波动：

```bash
./loadsim combo \
  --cpu-mode wave --cpu-min 20 --cpu-max 70 --cpu-period 60 \
  --ram-mode wave --ram-min-size 512 --ram-max-size 2048 --ram-period 90
```

## 命令说明

### `cpu`

```bash
./loadsim cpu [flags]
```

主要参数：

- `--mode`：`fixed` 或 `wave`
- `--percent`：固定模式 CPU 目标占用百分比
- `--min`：波动模式最小 CPU 百分比
- `--max`：波动模式最大 CPU 百分比
- `--period`：波动周期，单位秒
- `--cores`：worker 数，`0` 表示使用主机全部核心
- `--time`：运行时长，单位秒，`0` 表示不自动停止
- `--status-interval`：状态输出间隔，单位秒

帮助输出示例：

```text
$ ./loadsim cpu --help
Occupy CPU with fixed or wave patterns

Usage:
  loadsim cpu [flags]

Flags:
      --cores int             worker core count, 0 uses all host cores
      --max float             wave mode maximum CPU percent (default 80)
      --min float             wave mode minimum CPU percent (default 20)
      --mode string           fixed or wave (default "fixed")
      --percent float         fixed CPU target percent (default 50)
      --period int            wave mode period in seconds (default 60)
      --status-interval int   status print interval in seconds (default 2)
      --time int              run time in seconds, 0 means no limit
```

### `ram`

```bash
./loadsim ram [flags]
```

主要参数：

- `--mode`：`fixed` 或 `wave`
- `--size`：固定模式内存大小，单位 MB
- `--min-size`：波动模式最小内存，单位 MB
- `--max-size`：波动模式最大内存，单位 MB
- `--period`：波动周期，单位秒
- `--time`：运行时长，单位秒
- `--status-interval`：状态输出间隔，单位秒

帮助输出示例：

```text
$ ./loadsim ram --help
Occupy RAM with fixed or wave patterns

Usage:
  loadsim ram [flags]

Flags:
      --max-size int          wave mode maximum RAM in MB (default 1024)
      --min-size int          wave mode minimum RAM in MB (default 256)
      --mode string           fixed or wave (default "fixed")
      --period int            wave mode period in seconds (default 60)
      --size int              fixed RAM target in MB (default 1024)
      --status-interval int   status print interval in seconds (default 2)
      --time int              run time in seconds, 0 means no limit
```

### `combo`

```bash
./loadsim combo [flags]
```

主要参数分成两组：

CPU 相关：

- `--cpu-mode`
- `--cpu-percent`
- `--cpu-min`
- `--cpu-max`
- `--cpu-period`
- `--cpu-cores`

RAM 相关：

- `--ram-mode`
- `--ram-size`
- `--ram-min-size`
- `--ram-max-size`
- `--ram-period`

通用：

- `--time`
- `--status-interval`

帮助输出示例：

```text
$ ./loadsim combo --help
Occupy CPU and RAM at the same time

Usage:
  loadsim combo [flags]

Flags:
      --cpu-cores int         worker core count, 0 uses all host cores
      --cpu-max float         wave CPU maximum percent (default 80)
      --cpu-min float         wave CPU minimum percent (default 20)
      --cpu-mode string       CPU mode: fixed or wave (default "fixed")
      --cpu-percent float     fixed CPU target percent (default 50)
      --cpu-period int        wave CPU period in seconds (default 60)
      --ram-max-size int      wave RAM maximum in MB (default 1024)
      --ram-min-size int      wave RAM minimum in MB (default 256)
      --ram-mode string       RAM mode: fixed or wave (default "fixed")
      --ram-period int        wave RAM period in seconds (default 60)
      --ram-size int          fixed RAM target in MB (default 1024)
      --status-interval int   status print interval in seconds (default 2)
      --time int              run time in seconds, 0 means no limit
```

## 实际运行示例

下面这些都是直接运行得到的真实输出，第一次使用时可以直接对照。

### 示例 1：固定 CPU 占用

命令：

```bash
./loadsim cpu --mode fixed --percent 10 --cores 1 --time 1 --status-interval 1
```

输出：

```text
[09:39:56] cpu mode=fixed target=10.0% workers=1 host_cpu=36.7% host_mem=19.4%
[09:39:57] cpu mode=fixed target=10.0% workers=1 host_cpu=8.3% host_mem=19.6%
stopped: time limit reached
```

### 示例 2：固定 RAM 占用

命令：

```bash
./loadsim ram --mode fixed --size 64 --time 1 --status-interval 1
```

输出：

```text
[09:39:56] ram mode=fixed target=64MB current=64MB host_cpu=6.7% host_mem=19.6%
[09:39:57] ram mode=fixed target=64MB current=64MB host_cpu=5.0% host_mem=18.9%
stopped: time limit reached
```

### 示例 3：同时占用 CPU 和 RAM

命令：

```bash
./loadsim combo --cpu-percent 10 --cpu-cores 1 --ram-size 64 --time 1 --status-interval 1
```

输出：

```text
[09:39:56] combo cpu_target=10.0% workers=1 ram_target=64MB ram_current=64MB host_cpu=27.9% host_mem=19.6%
[09:39:57] combo cpu_target=10.0% workers=1 ram_target=64MB ram_current=64MB host_cpu=8.3% host_mem=19.4%
stopped: time limit reached
```

## 输出字段说明

### `cpu` 输出

```text
[09:39:56] cpu mode=fixed target=10.0% workers=1 host_cpu=36.7% host_mem=19.4%
```

字段解释：

- `mode`：当前模式
- `target`：当前 CPU 目标占用
- `workers`：当前 CPU worker 数
- `host_cpu`：主机实时 CPU 使用率
- `host_mem`：主机实时内存使用率

### `ram` 输出

```text
[09:39:56] ram mode=fixed target=64MB current=64MB host_cpu=6.7% host_mem=19.6%
```

字段解释：

- `target`：目标内存大小
- `current`：当前已占用内存大小

### `combo` 输出

```text
[09:39:56] combo cpu_target=10.0% workers=1 ram_target=64MB ram_current=64MB host_cpu=27.9% host_mem=19.6%
```

字段解释：

- `cpu_target`：当前 CPU 目标值
- `workers`：CPU worker 数
- `ram_target`：目标内存值
- `ram_current`：当前实际占用内存值
- `host_cpu`：主机当前 CPU 使用率
- `host_mem`：主机当前内存使用率

## 使用建议

### 固定场景

适合：

- 长时间保持相对稳定的资源占用
- 观察平稳图形
- 验证阈值附近的持续状态

推荐命令：

```bash
./loadsim cpu --mode fixed --percent 30 --cores 2
./loadsim ram --mode fixed --size 2048
./loadsim combo --cpu-percent 30 --cpu-cores 2 --ram-size 2048
```

### 波动场景

适合：

- 制造周期性曲线
- 观察画图效果
- 观察波峰波谷变化

推荐命令：

```bash
./loadsim cpu --mode wave --min 20 --max 70 --period 90 --cores 2
./loadsim ram --mode wave --min-size 512 --max-size 4096 --period 120
```

## 注意事项

- `cpu` 的百分比是针对所选 worker 集合的目标值，不是对整台机器的绝对精确硬限幅。
- `--cores=0` 表示使用主机全部核心。
- `ram` 会真实分配并触碰内存页，确保占用真正落到内存上。
- `host_cpu` 和 `host_mem` 是主机实时状态，会受到主机上其他进程影响。
- `--time 0` 表示不自动停止，需要手动 `Ctrl+C` 结束。
