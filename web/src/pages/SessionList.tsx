import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Tag, Modal, message, Drawer, Form, Select, Divider } from 'antd';
import { PlayCircle, StopCircle, Trash2, Terminal, Plus } from 'lucide-react';
import { api } from '../services/api';
import type { Session, Capability } from '../services/api';
import ConnectionInfo from '../components/ConnectionInfo';

const SessionList: React.FC = () => {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedSession, setSelectedSession] = useState<Session | null>(null);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [capabilities, setCapabilities] = useState<Capability[]>([]);
  const [form] = Form.useForm();

  const fetchSessions = async () => {
    setLoading(true);
    try {
      const data = await api.getSessions();
      setSessions(data);
    } catch (error) {
      message.error('Failed to fetch sessions');
    } finally {
      setLoading(false);
    }
  };

  const fetchCapabilities = async () => {
    try {
      const data = await api.getCapabilities();
      setCapabilities(data);
    } catch (error) {
      message.error('Failed to fetch capabilities');
    }
  };

  useEffect(() => {
    fetchSessions();
    fetchCapabilities();
    const interval = setInterval(fetchSessions, 5000); // Auto refresh
    return () => clearInterval(interval);
  }, []);

  const handlePower = async (id: string, action: 'poweron' | 'poweroff' | 'reset') => {
    try {
      if (action === 'poweron') await api.powerOn(id);
      else if (action === 'poweroff') await api.powerOff(id);
      else await api.reset(id);
      message.success(`Session ${action} successful`);
      fetchSessions();
    } catch (error) {
      message.error(`Failed to ${action} session`);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await api.deleteSession(id);
      message.success('Session deleted');
      fetchSessions();
    } catch (error) {
      message.error('Failed to delete session');
    }
  };

  const handleCreate = async (values: any) => {
    try {
      let boardConfig;
      if (values.template) {
        // Use custom template
        const templates = JSON.parse(localStorage.getItem('board_templates') || '[]');
        const template = templates.find((t: any) => t.name === values.template);
        if (template) {
          boardConfig = {
            ...template,
            board: 'custom' // Backend might need to know it's custom
          };
        }
      } else {
        // Use predefined board
        boardConfig = { board: values.board };
      }

      await api.createSession({
        backend: values.backend,
        board_config: boardConfig
      });
      message.success('Session created');
      setIsCreateModalOpen(false);
      fetchSessions();
    } catch (error) {
      message.error('Failed to create session');
    }
  };

  const [templates, setTemplates] = useState<any[]>([]);
  useEffect(() => {
    if (isCreateModalOpen) {
      const saved = JSON.parse(localStorage.getItem('board_templates') || '[]');
      setTemplates(saved);
    }
  }, [isCreateModalOpen]);

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      render: (text: string) => <Tag>{text.substring(0, 8)}...</Tag>,
    },
    {
      title: 'Backend',
      dataIndex: 'backend',
      key: 'backend',
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: 'Board',
      dataIndex: 'board',
      key: 'board',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        let color = 'default';
        if (status === 'running') color = 'success';
        if (status === 'error') color = 'error';
        return <Tag color={color}>{status.toUpperCase()}</Tag>;
      },
    },
    {
      title: 'GDB Port',
      dataIndex: 'gdb_port',
      key: 'gdb_port',
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: Session) => (
        <Space>
          <Button 
            icon={<PlayCircle size={16} />} 
            onClick={() => handlePower(record.id, 'poweron')}
            disabled={record.status === 'running'}
          />
          <Button 
            icon={<StopCircle size={16} />} 
            onClick={() => handlePower(record.id, 'poweroff')}
            disabled={record.status !== 'running'}
          />
          <Button 
            icon={<Terminal size={16} />} 
            onClick={() => setSelectedSession(record)}
          >
            Connect
          </Button>
          <Button 
            danger 
            icon={<Trash2 size={16} />} 
            onClick={() => handleDelete(record.id)}
          />
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>Simulation Sessions</h2>
        <Button type="primary" icon={<Plus size={16} />} onClick={() => setIsCreateModalOpen(true)}>
          New Session
        </Button>
      </div>

      <Table 
        columns={columns} 
        dataSource={sessions} 
        rowKey="id" 
        loading={loading} 
      />

      <Drawer
        title={`Connection: ${selectedSession?.id}`}
        placement="right"
        onClose={() => setSelectedSession(null)}
        open={!!selectedSession}
        width={500}
      >
        {selectedSession && <ConnectionInfo session={selectedSession} />}
      </Drawer>

      <Modal
        title="Create New Session"
        open={isCreateModalOpen}
        onCancel={() => setIsCreateModalOpen(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="backend" label="Backend" rules={[{ required: true }]}>
            <Select onChange={() => {
              form.setFieldValue('board', undefined);
              form.setFieldValue('template', undefined);
            }}>
              {capabilities.map(cap => (
                <Select.Option key={cap.backend} value={cap.backend}>{cap.backend}</Select.Option>
              ))}
            </Select>
          </Form.Item>
          
          <Form.Item 
            noStyle 
            shouldUpdate={(prev, curr) => prev.backend !== curr.backend}
          >
            {({ getFieldValue }) => {
              const backend = getFieldValue('backend');
              const boards = capabilities.find(c => c.backend === backend)?.boards || [];
              const backendTemplates = templates.filter(t => t.backend === backend);
              
              return (
                <>
                  <Divider>Predefined Boards</Divider>
                  <Form.Item name="board" label="Select Board">
                    <Select disabled={!backend} showSearch onChange={() => form.setFieldValue('template', undefined)}>
                      {boards.map(board => (
                        <Select.Option key={board} value={board}>{board}</Select.Option>
                      ))}
                    </Select>
                  </Form.Item>

                  {backendTemplates.length > 0 && (
                    <>
                      <Divider>Custom Templates</Divider>
                      <Form.Item name="template" label="Select Template">
                        <Select disabled={!backend} onChange={() => form.setFieldValue('board', undefined)}>
                          {backendTemplates.map((t: any) => (
                            <Select.Option key={t.name} value={t.name}>{t.name}</Select.Option>
                          ))}
                        </Select>
                      </Form.Item>
                    </>
                  )}
                </>
              );
            }}
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default SessionList;
