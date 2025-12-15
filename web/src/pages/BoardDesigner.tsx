import React, { useState, useEffect } from 'react';
import { 
  Form, Select, Input, Button, Card, Table, 
  message, Row, Col, Tabs, Typography 
} from 'antd';
import { Plus, Save, Trash2, MemoryStick, Box } from 'lucide-react';
import { api } from '../services/api';
import type { Capability } from '../services/api';

const { Title } = Typography;
const { Option } = Select;

interface MemoryRegion {
  key: string;
  name: string;
  start: string;
  size: string;
  permissions: string;
}

interface PeripheralConfig {
  key: string;
  name: string;
  type: string;
  address: string;
  properties: string; // JSON string for simplicity in MVP
}

const BoardDesigner: React.FC = () => {
  const [form] = Form.useForm();
  const [capabilities, setCapabilities] = useState<Capability[]>([]);
  const [selectedBackend, setSelectedBackend] = useState<string>('');
  
  // State for dynamic lists
  const [memoryRegions, setMemoryRegions] = useState<MemoryRegion[]>([]);
  const [peripherals, setPeripherals] = useState<PeripheralConfig[]>([]);

  useEffect(() => {
    loadCapabilities();
  }, []);

  const loadCapabilities = async () => {
    try {
      const data = await api.getCapabilities();
      setCapabilities(data);
    } catch (error) {
      message.error('Failed to load capabilities');
    }
  };

  const handleSave = (values: any) => {
    const config = {
      name: values.name,
      backend: values.backend,
      cpu: values.cpu,
      memory_map: memoryRegions.map(({ name, start, size, permissions }) => ({
        name, start, size, permissions
      })),
      peripherals: peripherals.map(({ name, type, address, properties }) => ({
        name, type, address, 
        properties: properties ? JSON.parse(properties || '{}') : {}
      }))
    };

    // Save to LocalStorage
    const savedTemplates = JSON.parse(localStorage.getItem('board_templates') || '[]');
    savedTemplates.push(config);
    localStorage.setItem('board_templates', JSON.stringify(savedTemplates));
    
    message.success('Board template saved successfully!');
  };

  // Memory Map Columns
  const memoryColumns = [
    { title: 'Name', dataIndex: 'name', render: (_: any, r: MemoryRegion) => 
      <Input value={r.name} onChange={e => updateMemory(r.key, 'name', e.target.value)} /> },
    { title: 'Start Addr (Hex)', dataIndex: 'start', render: (_: any, r: MemoryRegion) => 
      <Input value={r.start} onChange={e => updateMemory(r.key, 'start', e.target.value)} /> },
    { title: 'Size', dataIndex: 'size', render: (_: any, r: MemoryRegion) => 
      <Input value={r.size} onChange={e => updateMemory(r.key, 'size', e.target.value)} /> },
    { title: 'Perms', dataIndex: 'permissions', render: (_: any, r: MemoryRegion) => 
      <Select value={r.permissions} onChange={v => updateMemory(r.key, 'permissions', v)} style={{ width: 100 }}>
        <Option value="rwx">RWX</Option><Option value="rw">RW</Option><Option value="rx">RX</Option>
      </Select> },
    { title: 'Action', render: (_: any, r: MemoryRegion) => 
      <Button danger icon={<Trash2 size={14} />} onClick={() => setMemoryRegions(prev => prev.filter(i => i.key !== r.key))} /> }
  ];

  const updateMemory = (key: string, field: keyof MemoryRegion, value: any) => {
    setMemoryRegions(prev => prev.map(item => item.key === key ? { ...item, [field]: value } : item));
  };

  const addMemory = () => {
    setMemoryRegions([...memoryRegions, { 
      key: Date.now().toString(), name: 'RAM', start: '0x20000000', size: '0x10000', permissions: 'rwx' 
    }]);
  };

  // Peripheral Columns
  const peripheralColumns = [
    { title: 'Name', dataIndex: 'name', render: (_: any, r: PeripheralConfig) => 
      <Input value={r.name} onChange={e => updatePeripheral(r.key, 'name', e.target.value)} /> },
    { title: 'Type', dataIndex: 'type', render: (_: any, r: PeripheralConfig) => 
      <Select value={r.type} onChange={v => updatePeripheral(r.key, 'type', v)} style={{ width: 150 }}>
        {capabilities.find(c => c.backend === selectedBackend)?.peripherals.map(p => 
          <Option key={p} value={p}>{p}</Option>
        )}
      </Select> },
    { title: 'Address', dataIndex: 'address', render: (_: any, r: PeripheralConfig) => 
      <Input value={r.address} onChange={e => updatePeripheral(r.key, 'address', e.target.value)} /> },
    { title: 'Action', render: (_: any, r: PeripheralConfig) => 
      <Button danger icon={<Trash2 size={14} />} onClick={() => setPeripherals(prev => prev.filter(i => i.key !== r.key))} /> }
  ];

  const updatePeripheral = (key: string, field: keyof PeripheralConfig, value: any) => {
    setPeripherals(prev => prev.map(item => item.key === key ? { ...item, [field]: value } : item));
  };

  const addPeripheral = () => {
    setPeripherals([...peripherals, { 
      key: Date.now().toString(), name: 'UART0', type: 'uart', address: '0x40000000', properties: '{}' 
    }]);
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Title level={2}>Board Designer</Title>
        <Button type="primary" icon={<Save size={16} />} onClick={() => form.submit()}>
          Save Template
        </Button>
      </div>

      <Form form={form} layout="vertical" onFinish={handleSave}>
        <Row gutter={24}>
          <Col span={8}>
            <Card title="Basic Info">
              <Form.Item name="name" label="Template Name" rules={[{ required: true }]}>
                <Input placeholder="e.g., My Custom IoT Board" />
              </Form.Item>
              <Form.Item name="backend" label="Backend" rules={[{ required: true }]}>
                <Select onChange={setSelectedBackend}>
                  {capabilities.map(c => <Option key={c.backend} value={c.backend}>{c.backend}</Option>)}
                </Select>
              </Form.Item>
              <Form.Item name="cpu" label="Processor" rules={[{ required: true }]}>
                <Select disabled={!selectedBackend} showSearch>
                  {capabilities.find(c => c.backend === selectedBackend)?.processors.map(p => 
                    <Option key={p} value={p}>{p}</Option>
                  )}
                </Select>
              </Form.Item>
            </Card>
          </Col>
          
          <Col span={16}>
            <Card title="Hardware Configuration">
              <Tabs items={[
                {
                  key: 'memory',
                  label: <span><MemoryStick size={14} /> Memory Map</span>,
                  children: (
                    <>
                      <Table 
                        dataSource={memoryRegions} 
                        columns={memoryColumns} 
                        pagination={false} 
                        size="small" 
                      />
                      <Button type="dashed" onClick={addMemory} block icon={<Plus size={14} />} style={{ marginTop: 8 }}>
                        Add Memory Region
                      </Button>
                    </>
                  )
                },
                {
                  key: 'peripherals',
                  label: <span><Box size={14} /> Peripherals</span>,
                  children: (
                    <>
                      <Table 
                        dataSource={peripherals} 
                        columns={peripheralColumns} 
                        pagination={false} 
                        size="small" 
                      />
                      <Button type="dashed" onClick={addPeripheral} block icon={<Plus size={14} />} style={{ marginTop: 8 }}>
                        Add Peripheral
                      </Button>
                    </>
                  )
                }
              ]} />
            </Card>
          </Col>
        </Row>
      </Form>
    </div>
  );
};

export default BoardDesigner;
