import axios from 'axios';

const API_BASE_URL = '/api/v1';

export const client = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Types
export interface Session {
  id: string;
  backend: 'qemu' | 'renode';
  board: string;
  status: 'running' | 'stopped' | 'error' | 'created';
  gdb_port?: number;
  created_at: string;
  board_config?: any;
}

export interface Capability {
  backend: string;
  boards: string[];
  processors: string[];
  peripherals: string[];
  bus_types?: string[];
  features?: string[];
}

export interface CreateSessionRequest {
  backend: string;
  board_config: {
    board: string;
    [key: string]: any;
  };
}

// API Methods
export const api = {
  getCapabilities: async () => {
    const response = await client.get<Capability[]>('/capabilities');
    return response.data;
  },

  getSessions: async () => {
    const response = await client.get<Session[]>('/sessions');
    return response.data;
  },

  createSession: async (data: CreateSessionRequest) => {
    const response = await client.post<Session>('/sessions', data);
    return response.data;
  },

  deleteSession: async (id: string) => {
    await client.delete(`/sessions/${id}`);
  },

  powerOn: async (id: string) => {
    await client.post(`/sessions/${id}/poweron`);
  },

  powerOff: async (id: string) => {
    await client.post(`/sessions/${id}/poweroff`);
  },

  reset: async (id: string) => {
    await client.post(`/sessions/${id}/reset`);
  },
};
