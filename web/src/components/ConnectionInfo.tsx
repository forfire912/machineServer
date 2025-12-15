import React from 'react';
import { Card, Typography, Button, Space } from 'antd';
import { Copy } from 'lucide-react';
import type { Session } from '../services/api';

const { Text, Paragraph } = Typography;

interface ConnectionInfoProps {
  session: Session;
}

const ConnectionInfo: React.FC<ConnectionInfoProps> = ({ session }) => {
  if (!session.gdb_port) {
    return <Text type="secondary">Waiting for GDB port allocation...</Text>;
  }

  const launchConfig = {
    version: "0.2.0",
    configurations: [
      {
        name: `Debug ${session.board} (${session.backend})`,
        type: "cppdbg",
        request: "launch",
        program: "${workspaceFolder}/build/firmware.elf",
        MIMode: "gdb",
        miDebuggerPath: "/usr/bin/gdb-multiarch",
        miDebuggerServerAddress: `localhost:${session.gdb_port}`,
        cwd: "${workspaceFolder}",
        setupCommands: [
          {
            description: "Enable pretty-printing for gdb",
            text: "-enable-pretty-printing",
            ignoreFailures: true
          }
        ]
      }
    ]
  };

  const configString = JSON.stringify(launchConfig, null, 2);

  const handleCopy = () => {
    navigator.clipboard.writeText(configString);
  };

  return (
    <Space direction="vertical" style={{ width: '100%' }}>
      <Card size="small" title="Connection Details">
        <Space direction="vertical">
          <Text strong>GDB Port: <Text code>{session.gdb_port}</Text></Text>
          <Text strong>Host: <Text code>localhost</Text></Text>
        </Space>
      </Card>
      
      <Card size="small" title="VS Code Launch Configuration" extra={
        <Button type="text" icon={<Copy size={16} />} onClick={handleCopy}>
          Copy
        </Button>
      }>
        <Paragraph>
          <pre style={{ fontSize: '12px', maxHeight: '200px', overflow: 'auto' }}>
            {configString}
          </pre>
        </Paragraph>
      </Card>
    </Space>
  );
};

export default ConnectionInfo;
