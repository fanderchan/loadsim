# LoadSim

[![CI](https://github.com/fanderchan/loadsim/actions/workflows/ci.yml/badge.svg)](https://github.com/fanderchan/loadsim/actions/workflows/ci.yml)

`LoadSim` 是一个 Linux 负载场景模拟命令行工具，用于构造可控的 CPU、内存，以及 CPU+内存组合占用场景。

- 英文名：`LoadSim`
- 中文名：`负载场景模拟器`
- 命令名：`loadsim`

它适合做这些事情：

- 维持一段相对稳定的 CPU 占用
- 维持一段相对稳定的内存占用
- 让 CPU 或内存按周期波动
- 同时制造 CPU 和内存占用
- 验证监控、观察曲线、复现截图、做阈值演练

当前版本聚焦三条命令：

- `cpu`
- `ram`
- `combo`

## 核心概念

### CPU 有两种目标语义

`cpu` 和 `combo` 里的 CPU 部分都支持两种范围：

- `--scope workers`
  百分比作用在 worker 集合上。
- `--scope host`
  百分比作用在整机 CPU 上，LoadSim 会根据主机实时 CPU 使用率自适应调节自身负载。

例子：

```bash
./loadsim cpu --mode fixed --scope workers --percent 50 --cores 4
```

这条命令的含义不是整机 `50%`，而是 4 个 worker 总体按 `50%` 运行，大约相当于 `200%` 单核 CPU 负载。

如果机器有 `160` 核，那么整机只相当于约 `1.25%`。

如果你要的是“主机稳定占用 50% CPU，并在其他进程变忙时主动让出 CPU”，应该写成：

```bash
./loadsim cpu --mode fixed --scope host --percent 50
```

### CPU 有两种空闲策略

`cpu` 和 `combo` 里的 CPU 部分都支持：

- `--idle-mode park`
  预先保留 worker 池，空闲 worker 进入 sleep。线程/协程 churn 更低，适合长时间稳定占用，默认推荐。
- `--idle-mode trim`
  只保留当前需要的 worker，空闲 worker 会被回收。空闲 footprint 更小，但频繁波动时创建/销毁更频繁。

简单说：

- 想要更稳，优先用 `park`
- 想要更省空闲 worker，才用 `trim`

### CPU 自适应控制参数

当 `--scope host` 时，下面几个参数会影响稳定性：

- `--control-ms`
  控制器多久调整一次自身负载。
- `--sample-ms`
  每次取主机 CPU 样本的时间窗口。
- `--deadband`
  偏差在这个范围内时不调整，避免来回抖动。
- `--max-step`
  每次调整最多改多少 worker drive，避免一步拉太猛。

### RAM 平滑参数

`ram` 和 `combo` 里的 RAM 部分支持：

- `--block-size`
  内存分配块大小，越小越细腻，但块数更多。
- `--control-ms`
  内存控制周期。
- `--rate-limit`
  每秒最多调整多少 MB。它会限制启动阶段和运行阶段的内存变化速度，适合线上环境减小突刺。

## 编译

直接编译：

```bash
go build -o loadsim .
```

使用脚本编译：

```bash
./build.sh
```

默认构建产物：

```bash
build/loadsim
```

## 第一次使用

建议先看帮助和版本：

```bash
./loadsim --help
./loadsim cpu --help
./loadsim ram --help
./loadsim combo --help
./loadsim version
```

真实输出：

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

Flags:
  -h, --help   help for loadsim

Use "loadsim [command] --help" for more information about a command.
```

```text
$ ./loadsim version
LoadSim 0.3.0-beta.1
```

## 快速开始

### 1. 稳定占住整机 50% CPU

```bash
./loadsim cpu --mode fixed --scope host --percent 50
```

推荐先从默认参数开始：

- `--idle-mode park`
- `--control-ms 250`
- `--sample-ms 200`
- `--deadband 1`
- `--max-step 10`

### 2. 固定制造 200% 单核等效 CPU 负载

```bash
./loadsim cpu --mode fixed --scope workers --percent 50 --cores 4
```

这表示 4 个 worker 共同跑出约 `200%` 单核 CPU 负载，不是整机 `50%`。

### 3. 固定占用 1GB 内存

```bash
./loadsim ram --mode fixed --size 1024
```

### 4. 同时占用 CPU 和内存

```bash
./loadsim combo --cpu-scope host --cpu-percent 30 --ram-size 1024
```

### 5. 跑一个更平滑的小内存波动

```bash
./loadsim ram \
  --mode wave \
  --min-size 64 \
  --max-size 256 \
  --period 120 \
  --block-size 8 \
  --rate-limit 16
```

## 命令说明

### `cpu`

```bash
./loadsim cpu [flags]
```

常用参数：

- `--mode fixed|wave`
- `--scope workers|host`
- `--idle-mode park|trim`
- `--percent`
- `--min`
- `--max`
- `--period`
- `--cores`
- `--control-ms`
- `--sample-ms`
- `--deadband`
- `--max-step`
- `--time`
- `--status-interval`

真实帮助输出：

```text
$ ./loadsim cpu --help
Occupy CPU with fixed or wave patterns

Usage:
  loadsim cpu [flags]

Flags:
      --control-ms int        controller adjustment interval in milliseconds (default 250)
      --cores int             worker core count, 0 uses all host cores
      --deadband float        host CPU deadband percent before adjusting (default 1)
  -h, --help                  help for cpu
      --idle-mode string      idle worker behavior: park or trim (default "park")
      --max float             wave mode maximum CPU percent (default 80)
      --max-step float        maximum worker drive change per control step in percent (default 10)
      --min float             wave mode minimum CPU percent (default 20)
      --mode string           fixed or wave (default "fixed")
      --percent float         fixed CPU target percent (default 50)
      --period int            wave mode period in seconds (default 60)
      --sample-ms int         host CPU sample duration in milliseconds (default 200)
      --scope string          CPU target scope: workers or host (default "workers")
      --status-interval int   status print interval in seconds (default 2)
      --time int              run time in seconds, 0 means no limit
```

### `ram`

```bash
./loadsim ram [flags]
```

常用参数：

- `--mode fixed|wave`
- `--size`
- `--min-size`
- `--max-size`
- `--period`
- `--block-size`
- `--control-ms`
- `--rate-limit`
- `--time`
- `--status-interval`

真实帮助输出：

```text
$ ./loadsim ram --help
Occupy RAM with fixed or wave patterns

Usage:
  loadsim ram [flags]

Flags:
      --block-size int        RAM allocation block size in MB (default 16)
      --control-ms int        RAM control interval in milliseconds (default 250)
  -h, --help                  help for ram
      --max-size int          wave mode maximum RAM in MB (default 1024)
      --min-size int          wave mode minimum RAM in MB (default 256)
      --mode string           fixed or wave (default "fixed")
      --period int            wave mode period in seconds (default 60)
      --rate-limit int        RAM change rate limit in MB per second, 0 means unlimited
      --size int              fixed RAM target in MB (default 1024)
      --status-interval int   status print interval in seconds (default 2)
      --time int              run time in seconds, 0 means no limit
```

### `combo`

```bash
./loadsim combo [flags]
```

CPU 相关参数：

- `--cpu-mode`
- `--cpu-scope`
- `--cpu-idle-mode`
- `--cpu-percent`
- `--cpu-min`
- `--cpu-max`
- `--cpu-period`
- `--cpu-cores`
- `--cpu-control-ms`
- `--cpu-sample-ms`
- `--cpu-deadband`
- `--cpu-max-step`

RAM 相关参数：

- `--ram-mode`
- `--ram-size`
- `--ram-min-size`
- `--ram-max-size`
- `--ram-period`
- `--ram-block-size`
- `--ram-control-ms`
- `--ram-rate-limit`

真实帮助输出：

```text
$ ./loadsim combo --help
Occupy CPU and RAM at the same time

Usage:
  loadsim combo [flags]

Flags:
      --cpu-control-ms int     CPU controller adjustment interval in milliseconds (default 250)
      --cpu-cores int          worker core count, 0 uses all host cores
      --cpu-deadband float     CPU host deadband percent before adjusting (default 1)
      --cpu-idle-mode string   CPU idle worker behavior: park or trim (default "park")
      --cpu-max float          wave CPU maximum percent (default 80)
      --cpu-max-step float     maximum CPU worker drive change per control step in percent (default 10)
      --cpu-min float          wave CPU minimum percent (default 20)
      --cpu-mode string        CPU mode: fixed or wave (default "fixed")
      --cpu-percent float      fixed CPU target percent (default 50)
      --cpu-period int         wave CPU period in seconds (default 60)
      --cpu-sample-ms int      CPU host sample duration in milliseconds (default 200)
      --cpu-scope string       CPU target scope: workers or host (default "workers")
  -h, --help                   help for combo
      --ram-block-size int     RAM allocation block size in MB (default 16)
      --ram-control-ms int     RAM control interval in milliseconds (default 250)
      --ram-max-size int       wave RAM maximum in MB (default 1024)
      --ram-min-size int       wave RAM minimum in MB (default 256)
      --ram-mode string        RAM mode: fixed or wave (default "fixed")
      --ram-period int         wave RAM period in seconds (default 60)
      --ram-rate-limit int     RAM change rate limit in MB per second, 0 means unlimited
      --ram-size int           fixed RAM target in MB (default 1024)
      --status-interval int    status print interval in seconds (default 2)
      --time int               run time in seconds, 0 means no limit
```

## 实际运行示例

下面这些都是命令直接运行得到的真实输出，第一次使用时可以直接照着对比。

### 示例 1：固定 worker 范围 CPU 占用

命令：

```bash
./loadsim cpu --mode fixed --scope workers --percent 10 --cores 1 --time 1 --status-interval 1
```

输出：

```text
[18:33:02] cpu mode=fixed scope=workers idle=park target=10.0% drive=10.0% workers=1/1 host_cpu=3.4% host_mem=19.5%
[18:33:03] cpu mode=fixed scope=workers idle=park target=10.0% drive=10.0% workers=1/1 host_cpu=10.3% host_mem=19.4%
stopped: time limit reached
```

### 示例 2：整机 CPU 自适应占用

命令：

```bash
./loadsim cpu --mode fixed --scope host --percent 20 --time 3 --status-interval 1
```

输出：

```text
[18:33:21] cpu mode=fixed scope=host idle=park target=20.0% drive=17.5% workers=1/4 host_cpu=41.0% host_mem=19.3%
[18:33:22] cpu mode=fixed scope=host idle=park target=20.0% drive=6.2% workers=1/4 host_cpu=13.6% host_mem=19.3%
[18:33:23] cpu mode=fixed scope=host idle=park target=20.0% drive=6.0% workers=1/4 host_cpu=14.8% host_mem=19.3%
[18:33:24] cpu mode=fixed scope=host idle=park target=20.0% drive=7.6% workers=1/4 host_cpu=16.4% host_mem=19.3%
stopped: time limit reached
```

这里的关键点是：

- `target` 是整机 CPU 目标值
- `drive` 是当前分配给 LoadSim 自己 worker 集合的实际驱动力度
- `workers=1/4` 表示当前 4 个 worker 预算里，只有 1 个处于活跃状态
- 如果其他进程变忙，`drive` 会下降

### 示例 3：固定内存占用

命令：

```bash
./loadsim ram --mode fixed --size 64 --block-size 8 --time 1 --status-interval 1
```

输出：

```text
[18:33:01] ram mode=fixed target=64MB current=64MB block=8MB rate_limit=0MB/s host_cpu=8.3% host_mem=19.5%
[18:33:02] ram mode=fixed target=64MB current=64MB block=8MB rate_limit=0MB/s host_cpu=19.7% host_mem=20.0%
stopped: time limit reached
```

### 示例 4：带速率限制的 RAM 平滑增长

命令：

```bash
./loadsim ram --mode fixed --size 32 --block-size 4 --rate-limit 8 --time 3 --status-interval 1
```

输出：

```text
[18:35:00] ram mode=fixed target=2MB current=2MB block=4MB rate_limit=8MB/s host_cpu=1.6% host_mem=18.8%
[18:35:01] ram mode=fixed target=10MB current=10MB block=4MB rate_limit=8MB/s host_cpu=0.0% host_mem=18.8%
[18:35:02] ram mode=fixed target=18MB current=18MB block=4MB rate_limit=8MB/s host_cpu=6.6% host_mem=19.1%
[18:35:03] ram mode=fixed target=24MB current=24MB block=4MB rate_limit=8MB/s host_cpu=0.0% host_mem=19.1%
stopped: time limit reached
```

这个例子能看出 `--rate-limit` 不只是限制波动场景，也会限制固定目标在启动阶段的增长速度。

### 示例 5：组合场景

命令：

```bash
./loadsim combo --cpu-scope host --cpu-percent 20 --cpu-idle-mode park --ram-size 64 --ram-block-size 8 --time 3 --status-interval 1
```

输出：

```text
[18:33:21] combo cpu_scope=host cpu_idle=park cpu_target=20.0% cpu_drive=17.5% workers=1/4 ram_target=64MB ram_current=64MB host_cpu=38.7% host_mem=19.3%
[18:33:22] combo cpu_scope=host cpu_idle=park cpu_target=20.0% cpu_drive=4.3% workers=1/4 ram_target=64MB ram_current=64MB host_cpu=11.9% host_mem=19.3%
[18:33:23] combo cpu_scope=host cpu_idle=park cpu_target=20.0% cpu_drive=4.0% workers=1/4 ram_target=64MB ram_current=64MB host_cpu=14.8% host_mem=19.3%
[18:33:24] combo cpu_scope=host cpu_idle=park cpu_target=20.0% cpu_drive=4.9% workers=1/4 ram_target=64MB ram_current=64MB host_cpu=18.0% host_mem=19.3%
stopped: time limit reached
```

## 输出字段说明

### `cpu`

```text
[18:33:21] cpu mode=fixed scope=host idle=park target=20.0% drive=6.2% workers=1/4 host_cpu=13.6% host_mem=19.3%
```

- `mode`：当前模式
- `scope`：`workers` 或 `host`
- `idle`：`park` 或 `trim`
- `target`：当前 CPU 目标值
- `drive`：当前施加到 LoadSim worker 集合上的驱动值
- `workers`：当前活跃 worker 数 / 最大 worker 预算
- `host_cpu`：主机实时 CPU 使用率
- `host_mem`：主机实时内存使用率

### `ram`

```text
[18:35:01] ram mode=fixed target=10MB current=10MB block=4MB rate_limit=8MB/s host_cpu=0.0% host_mem=18.8%
```

- `target`：当前实际目标内存值
- `current`：当前实际已占用内存值
- `block`：内存分配块大小
- `rate_limit`：当前速率限制

### `combo`

```text
[18:33:22] combo cpu_scope=host cpu_idle=park cpu_target=20.0% cpu_drive=4.3% workers=1/4 ram_target=64MB ram_current=64MB host_cpu=11.9% host_mem=19.3%
```

- `cpu_scope`：CPU 目标作用范围
- `cpu_idle`：CPU 空闲策略
- `cpu_target`：当前 CPU 目标值
- `cpu_drive`：当前 CPU 实际驱动值
- `workers`：当前活跃 CPU worker 数 / 最大 worker 预算
- `ram_target`：当前 RAM 目标值
- `ram_current`：当前 RAM 实际占用值

## 生产调优建议

### 想让整机 CPU 更稳

推荐起点：

```bash
./loadsim cpu \
  --mode fixed \
  --scope host \
  --percent 50 \
  --idle-mode park \
  --control-ms 250 \
  --sample-ms 300 \
  --deadband 1 \
  --max-step 5
```

建议：

- 长时间稳定占用，优先用 `park`
- 主机很吵、背景进程很多时，增大 `--sample-ms` 到 `300-800`
- 目标附近来回晃时，适当增大 `--deadband` 到 `1.5-2`
- 调整太猛时，降低 `--max-step` 到 `3-5`
- 反应太慢时，再把 `--max-step` 提到 `10-15` 或把 `--control-ms` 降到 `150-200`

### 想让 RAM 小目标和高频波动更平滑

推荐起点：

```bash
./loadsim ram \
  --mode wave \
  --min-size 64 \
  --max-size 256 \
  --period 120 \
  --block-size 4 \
  --control-ms 250 \
  --rate-limit 16
```

建议：

- 小目标内存优先用 `--block-size 4` 或 `8`
- 频繁波动时一定要考虑 `--rate-limit`
- `--period` 很短但振幅很大时，内存 churn 会明显增加
- 如果只是做稳定占用，`--block-size 16` 或 `32` 就够了

### 什么时候用 `trim`

如果你更关心“空闲时尽量不保留 worker”，可以用：

```bash
./loadsim cpu --mode fixed --scope host --percent 30 --idle-mode trim
```

但要知道它的代价：

- worker 创建和回收更频繁
- 高频波动时更容易引入额外抖动
- 一般不如 `park` 适合长时间稳定占用

## 注意事项

- `cpu --scope workers` 的百分比不是整机百分比。
- `cpu --scope host` 会按整机 CPU 百分比做自适应控制。
- `combo --cpu-scope host` 会让组合模式里的 CPU 部分也按整机 CPU 百分比做自适应控制。
- 在 `scope=host` 下，如果目标超出当前 worker 数能提供的整机上限，程序会直接报错。
- `--cores=0` 表示使用主机全部核心。
- `ram` 会真实分配并触碰内存页，确保占用真正落到内存上。
- `host_cpu` 和 `host_mem` 是主机实时状态，会受到主机其他进程影响。
- `--time 0` 表示不自动停止，需要手动 `Ctrl+C` 结束。
