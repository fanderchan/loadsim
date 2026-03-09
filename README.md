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
- `--scope` 默认为 `workers`
- 以上命令表示 4 个 worker 的总负载按 50% 运行，不是整机 50%
- 一直运行，直到手动停止

如果只运行 60 秒：

```bash
./loadsim cpu --mode fixed --percent 50 --cores 4 --time 60
```

如果你要的是“整机稳定占用 50% CPU，并在其他进程变忙时自动让出 CPU”，应使用：

```bash
./loadsim cpu --mode fixed --scope host --percent 50
```

这个模式会按整机 CPU 百分比做自适应控制：

- 如果主机当前 CPU 低于目标值，LoadSim 会增加自身负载
- 如果主机当前 CPU 接近或高于目标值，LoadSim 会降低自身负载
- 如果其他进程已经把主机 CPU 打到目标值以上，LoadSim 会把自身 CPU 占用降到接近 0

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
- `--scope`：`workers` 或 `host`
- `--percent`：固定模式 CPU 目标占用百分比
- `--min`：波动模式最小 CPU 百分比
- `--max`：波动模式最大 CPU 百分比
- `--period`：波动周期，单位秒
- `--cores`：worker 数，`0` 表示使用主机全部核心
- `--time`：运行时长，单位秒，`0` 表示不自动停止
- `--status-interval`：状态输出间隔，单位秒

`cpu` 有两种百分比语义：

- `--scope workers`：百分比作用在 worker 集合上
- `--scope host`：百分比作用在整机 CPU 上，并通过自适应控制来逼近目标值

例如在一台 160 核机器上：

```bash
./loadsim cpu --mode fixed --percent 50 --cores 4
```

这不是整机 50%，而是大约 `200%` 单核 CPU 负载，总体只相当于整机约 `1.25%`。

如果你写：

```bash
./loadsim cpu --mode fixed --scope host --percent 50 --cores 4
```

程序会直接报错，因为 4 个 worker 在 160 核机器上最多只能提供约 `2.5%` 的整机 CPU，目标不可达。

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
      --scope string          CPU target scope: workers or host (default "workers")
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
./loadsim cpu --mode fixed --scope workers --percent 10 --cores 1 --time 1 --status-interval 1
```

输出：

```text
[18:04:23] cpu mode=fixed scope=workers target=10.0% drive=10.0% workers=1 host_cpu=18.3% host_mem=19.3%
[18:04:24] cpu mode=fixed scope=workers target=10.0% drive=10.0% workers=1 host_cpu=9.8% host_mem=19.3%
stopped: time limit reached
```

### 示例 2：整机 CPU 自适应占用

命令：

```bash
./loadsim cpu --mode fixed --scope host --percent 10 --time 1 --status-interval 1
```

输出：

```text
[18:04:59] cpu mode=fixed scope=host target=10.0% drive=0.0% workers=4 host_cpu=0.0% host_mem=18.7%
[18:05:00] cpu mode=fixed scope=host target=10.0% drive=10.0% workers=4 host_cpu=18.3% host_mem=18.7%
stopped: time limit reached
```

这里 `scope=host` 的含义是：

- `target=10.0%` 是整机 CPU 目标值
- `drive=10.0%` 是当前施加到 worker 集合上的占用比例
- 当其他进程变忙时，`drive` 会下降；如果其他进程已经把整机 CPU 打到目标值以上，`drive` 会降到接近 `0`

### 示例 3：固定 RAM 占用

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

### 示例 4：同时占用 CPU 和 RAM

命令：

```bash
./loadsim combo --cpu-percent 10 --cpu-cores 1 --ram-size 64 --time 1 --status-interval 1
```

输出：

```text
[18:04:23] combo cpu_scope=workers cpu_target=10.0% cpu_drive=10.0% workers=1 ram_target=64MB ram_current=64MB host_cpu=5.1% host_mem=19.3%
[18:04:24] combo cpu_scope=workers cpu_target=10.0% cpu_drive=10.0% workers=1 ram_target=64MB ram_current=64MB host_cpu=8.2% host_mem=19.2%
stopped: time limit reached
```

## 输出字段说明

### `cpu` 输出

```text
[09:39:56] cpu mode=fixed scope=workers target=10.0% drive=10.0% workers=1 host_cpu=36.7% host_mem=19.4%
```

字段解释：

- `mode`：当前模式
- `scope`：当前 CPU 百分比的作用范围
- `target`：当前 CPU 目标值
- `drive`：当前实际施加到 worker 集合上的占用比例
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
[18:04:23] combo cpu_scope=workers cpu_target=10.0% cpu_drive=10.0% workers=1 ram_target=64MB ram_current=64MB host_cpu=5.1% host_mem=19.3%
```

字段解释：

- `cpu_scope`：CPU 百分比的作用范围
- `cpu_target`：当前 CPU 目标值
- `cpu_drive`：当前施加到 CPU worker 集合上的占用比例
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

- `cpu --scope workers` 的百分比是针对所选 worker 集合的目标值。
- `cpu --scope host` 会按整机 CPU 百分比做自适应控制。
- 在 `--scope host` 下，如果目标超出当前 worker 数可提供的整机上限，程序会直接报错。
- `--cores=0` 表示使用主机全部核心。
- `ram` 会真实分配并触碰内存页，确保占用真正落到内存上。
- `host_cpu` 和 `host_mem` 是主机实时状态，会受到主机上其他进程影响。
- `--time 0` 表示不自动停止，需要手动 `Ctrl+C` 结束。
