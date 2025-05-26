


![](doc/os服务化.drawio.png)

# 系统配置 API 设计

本目录包含了 nix-operator 的业务化 API 设计，分为两个主要的 proto 文件：

## 文件结构

- `os.proto` - 操作系统配置相关接口
- `hardware.proto` - 硬件配置相关接口  
- `schema-examples.json` - JSON Schema 示例，用于前端表单生成

## 设计理念

### 1. 业务化接口设计

将底层的系统配置抽象为业务友好的 API 接口，支持：

- **RESTful 风格**：每个配置模块都有独立的 Get/Update 接口
- **类型安全**：使用 protobuf 强类型定义，避免配置错误
- **状态查询**：不仅支持配置更新，还能查询当前状态
- **错误处理**：统一的错误响应格式

### 2. 前端表单生成支持

通过 JSON Schema 支持自动表单生成：

- **字段验证**：内置正则表达式和范围验证
- **UI 提示**：支持 placeholder、description 等 UI 元素
- **条件显示**：支持依赖关系和条件显示逻辑
- **本地化**：中文字段名称和描述

## API 模块说明

### OS 配置模块 (os.proto)

#### 网络配置
- **接口管理**：支持多网卡配置，IPv4/IPv6 双栈
- **节点选择**：通过主机名、MAC 地址等选择目标节点
- **DHCP 支持**：静态 IP 和 DHCP 自动获取
- **状态监控**：接口状态实时查询

#### DNS 配置
- **多服务器**：支持多个 DNS 服务器配置
- **搜索域**：支持域名搜索路径配置
- **选项配置**：支持 DNS 解析选项

#### Hosts 配置
- **批量管理**：支持批量 hosts 条目管理
- **注释支持**：每个条目可添加注释说明
- **多主机名**：单个 IP 可对应多个主机名

#### 时间配置
- **时区设置**：支持标准时区配置
- **NTP 同步**：支持多 NTP 服务器配置
- **状态监控**：NTP 同步状态实时查询
- **硬件时钟**：支持硬件时钟同步配置

#### 防火墙配置
- **规则管理**：支持端口、协议、IP 地址规则
- **默认策略**：支持默认允许/拒绝策略
- **规则描述**：每个规则可添加描述信息

### 硬件配置模块 (hardware.proto)

#### 串口配置
- **参数设置**：波特率、数据位、停止位、校验位
- **模式切换**：RS232/RS485/RS422 模式支持
- **RS485 配置**：RTS 控制、延迟设置、超时配置
- **透传功能**：TCP/UDP/WebSocket 透传支持
- **状态监控**：串口状态和透传状态查询
- **设备发现**：自动扫描和识别串口设备

#### Udev 规则
- **设备匹配**：基于供应商 ID、产品 ID 等属性匹配
- **符号链接**：创建固定的设备符号链接
- **权限设置**：设备文件权限和所有者配置
- **规则优先级**：支持规则优先级排序

#### 硬件扫描
- **设备发现**：自动扫描系统硬件设备
- **分类管理**：按设备类型分类显示
- **详细信息**：设备制造商、型号、ID 等详细信息
- **层次结构**：支持父子设备关系显示

## JSON Schema 表单生成

### 表单字段类型

```json
{
  "string": "文本输入框",
  "integer": "数字输入框", 
  "boolean": "复选框",
  "enum": "下拉选择框",
  "array": "动态列表"
}
```

### UI 控件支持

```json
{
  "ui:widget": "select|radio|checkbox|updown|textarea",
  "ui:placeholder": "输入提示文本",
  "ui:description": "字段说明",
  "ui:collapsible": "可折叠区域"
}
```

### 验证规则

```json
{
  "pattern": "正则表达式验证",
  "minimum/maximum": "数值范围验证", 
  "minItems/maxItems": "数组长度验证",
  "required": "必填字段验证"
}
```

## 使用示例

### 1. 生成 Go 代码

```bash
protoc --go_out=. --go-grpc_out=. api/system/v1/*.proto
```

### 2. 生成 JSON Schema

```bash
protoc --jsonschema_out=. api/system/v1/*.proto
```

### 3. 前端表单集成

```javascript
import { JSONSchemaForm } from '@rjsf/core';

// 使用生成的 schema
const schema = require('./generated/os_config_schema.json');
const uiSchema = require('./generated/ui_schema.json');

<JSONSchemaForm 
  schema={schema}
  uiSchema={uiSchema}
  onSubmit={handleSubmit}
/>
```

## 扩展性设计

### 1. 版本管理
- 使用 `v1` 版本目录，支持 API 版本演进
- 向后兼容的字段添加和废弃机制

### 2. 插件化
- 每个配置模块独立设计，支持插件化扩展
- 统一的响应格式和错误处理

### 3. 多平台支持
- 抽象的配置接口，底层可适配不同操作系统
- 平台特定的实现通过 Handler 模式隔离

## 最佳实践

1. **配置验证**：在 API 层进行配置验证，避免无效配置传递到系统层
2. **原子操作**：配置更新使用事务机制，确保一致性
3. **状态同步**：定期同步配置状态，确保配置和实际状态一致
4. **错误恢复**：配置失败时提供回滚机制
5. **审计日志**：记录所有配置变更操作，支持审计追踪 
