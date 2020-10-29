# Table of Contents

- [Installing](#installing)
- [Getting Started](#getting-started)
- [Features](#features)
  - [build-in command](#build-in-command)
    - [@echo](#echo)
    - [@var](#var)
    - [@env](#env)
    - [comment](#comment)
    - [@touch](#touch)
    - [@mv](#mv)
    - [@copy](#copy)
    - [@rm](#rm)
    - [@mkdir](#mkdir)
    - [@cd](#cd)
  - [system command](#system-command)

# Installing

使用 `go get` 来安装最新版本，这个命令将安装 gmake 可执行文件到\$GOPATH/bin, 并下载相关依赖

```sh
go get -u github.com/fdxxw/gmake
```

# Getting Started

在当前目录编写 gmake.yml 文件，内容如下

```yml
vars:
  msg: Hello World

all: |
  @echo {{.msg}}
```

之后在当前命令行控制台运行 `gmake`, 可以看到控制台打印

```
@Echo: Hello World
```

也可以通过 `gmake -c gmake.yml` 指定文件运行

通过 `gmake -h` 或 `gmake --help` 可以查看相关帮助

# Features

## build-in command

内置有如下命令

### @echo

打印信息

```sh
@echo msg
```

### @var

设置变量

```sh
@var msg Hello World
```

### @env

设置环境变量

```
@env GOOS linux
```

### comment

`#`开头的为注释

```
# comment
```

### @touch

创建文件

```
@touch from.txt
```

### @mv

移动文件或目录

```
@mv from.txt to.txt
```

### @copy

复制文件或目录

```
@copy to.txt from.txt
```

### @rm

删除文件或目录

```
@rm from.txt
```

### @mkdir

创建目录

```
@mkdir from
```

### @cd

设置目录，使后续控制台命令在指定目录运行,只对系统命令有效，对内置命令无效

```
@cd from
```

## system command

系统命令,执行控制台命令，控制台能执行的都能执行

```sh
go build
```

# examples

examples.yml

```yml
vars:
  msg: Hello World
all: |
  @echo {{.msg}}
  # 修改msg变量
  @var msg Hello
  @echo {{.msg}}
  # 创建文件
  @touch from.txt
  @mv from.txt to.txt
  @copy to.txt from.txt
  @rm from.txt
  @rm to.txt
  @mkdir from
  @mv from to
  @copy to from
  @rm from
  @rm to
  @env GOOS linux
  go build
```

```sh
gmake -c examples.yml
```
