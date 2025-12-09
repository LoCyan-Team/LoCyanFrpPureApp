#!/bin/bash
set -e

# 1. 编译当前平台版本用于获取版本号 (假设根目录原本的 Makefile 可以构建出 bin/frps)
# 如果你的主 Makefile 输出在 release 目录，请调整此处获取版本号的方式
echo "Building local version to get version number..."
go build -o ./bin/frps ./cmd/frps
if [ $? -ne 0 ]; then
    echo "make error"
    exit 1
fi

frp_version=`./bin/frps --version`
echo "build version: $frp_version"

# 2. 执行交叉编译 (调用上一轮优化过的 Makefile)
echo "Start cross-compiling..."
make -f ./Makefile.cross-compiles
if [ $? -ne 0 ]; then
    echo "Cross-compile failed!"
    exit 1
fi

# 3. 准备打包目录
rm -rf ./release/packages
mkdir -p ./release/packages

# 定义平台列表 (需与 Makefile 中的逻辑对应)
os_all='linux windows darwin freebsd android'
arch_all='386 amd64 arm arm64 mips64 mips64le mips mipsle riscv64 loong64'
extra_all='_ hf'

cd ./release

for os in $os_all; do
    for arch in $arch_all; do
        for extra in $extra_all; do
            # --- A. 构建后缀名 (Suffix) ---
            # 逻辑需与 Makefile 中生成 SUFFIX 的逻辑一致
            suffix="${os}_${arch}"
            if [ "x${extra}" != x"_" ]; then
                suffix="${os}_${arch}_${extra}"
            fi
            
            # 定义包名
            frp_dir_name="frp_${frp_version}_${suffix}"
            frp_path="./packages/${frp_dir_name}"

            # --- B. 确定源文件名和目标文件名 ---
            frpc_src=""
            frps_src=""
            frpc_dest="frpc"
            frps_dest="frps"
            
            # 根据 OS 特性调整文件名查找策略
            if [ "x${os}" = x"windows" ]; then
                # Windows: source has .exe, dest has .exe
                frpc_src="frpc_${suffix}.exe"
                frps_src="frps_${suffix}.exe"
                frpc_dest="frpc.exe"
                frps_dest="frps.exe"
            elif [ "x${os}" = x"android" ]; then
                # Android: Makefile 生成的是 libfrpc_... .so
                # 注意：Android 只构建 frpc，没有 frps
                frpc_src="libfrpc_${suffix}.so"
                frps_src="libfrps_${suffix}.so" # 通常不存在
                frpc_dest="libfrpc.so"          # 保持库文件命名习惯，或者改为 frpc.so
                frps_dest="libfrps.so"
            else
                # Linux / Darwin / FreeBSD
                frpc_src="frpc_${suffix}"
                frps_src="frps_${suffix}"
            fi

            # --- C. 检查核心文件是否存在 ---
            # 如果连 frpc 都没有，说明该架构未构建或跳过，直接 continue
            if [ ! -f "./${frpc_src}" ]; then
                continue
            fi

            # --- D. 开始打包 ---
            echo "Packaging ${frp_dir_name}..."
            mkdir -p ${frp_path}

            # 移动 frpc
            mv "./${frpc_src}" "${frp_path}/${frpc_dest}"

            # 移动 frps (如果存在)
            # 注意：Android 构建不包含 frps，所以这里要做判断，不能因为缺少 frps 就报错退出
            if [ -f "./${frps_src}" ]; then
                mv "./${frps_src}" "${frp_path}/${frps_dest}"
            fi

            # 复制配置文件和文档
            if [ -f "../LICENSE" ]; then
                cp ../LICENSE ${frp_path}
            fi
            if [ -f "../conf/frpc.toml" ]; then
                cp -f ../conf/frpc.toml ${frp_path}/frpc.toml
            fi
            if [ -f "../conf/frps.toml" ]; then
                cp -f ../conf/frps.toml ${frp_path}/frps.toml
            fi

            # --- E. 压缩归档 ---
            # 进入 packages 目录进行压缩，避免压缩包包含路径前缀
            cd ./packages
            if [ "x${os}" = x"windows" ]; then
                zip -rq ${frp_dir_name}.zip ${frp_dir_name}
            else
                tar -zcf ${frp_dir_name}.tar.gz ${frp_dir_name}
            fi
            
            # 清理临时目录
            rm -rf ${frp_dir_name}
            
            # 返回 release 目录继续循环
            cd ..
        done
    done
done

echo "All packages generated in release/packages/"
cd -
