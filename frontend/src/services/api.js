import axios from 'axios';

const API_BASE = '/api';

const api = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
});

// Attach token to every request
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('sm_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 globally
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('sm_token');
      localStorage.removeItem('sm_user');
      window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);

// Auth
export const authAPI = {
  register: (data) => api.post('/auth/register', data),
  login: (data) => api.post('/auth/login', data),
  logout: () => api.post('/auth/logout'),
  getProfile: () => api.get('/auth/profile'),
};

// Meetings
export const meetingAPI = {
  create: (data) => api.post('/meetings', data),
  get: (id) => api.get(`/meetings/${id}`),
  start: (id) => api.post(`/meetings/${id}/start`),
  join: (data) => api.post('/meetings/join', data),
  end: (id) => api.post(`/meetings/${id}/end`),
  getParticipants: (id) => api.get(`/meetings/${id}/participants`),
  updateParticipant: (id, data) => api.patch(`/meetings/${id}/participant`, data),
  kickParticipant: (id, userId) => api.delete(`/meetings/${id}/participants/${userId}`),
};

// Chat
export const chatAPI = {
  sendText: (data) => api.post('/chat/messages/text', data),
  sendFile: (data) => api.post('/chat/messages/file', data),
  sendVoice: (data) => api.post('/chat/messages/voice', data),
  sendPoll: (data) => api.post('/chat/messages/poll', data),
  votePoll: (data) => api.post('/chat/messages/poll/vote', data),
  addReaction: (data) => api.post('/chat/messages/reaction', data),
  removeReaction: (data) => api.delete('/chat/messages/reaction', { data }),
  editMessage: (data) => api.patch('/chat/messages/edit', data),
  deleteMessage: (data) => api.delete('/chat/messages', { data }),
  getMessages: (roomId, params) => api.get(`/chat/rooms/${roomId}/messages`, { params }),
  createRoom: (data) => api.post('/chat/rooms', data),
  getRooms: () => api.get('/chat/rooms'),
};

// Bandwidth
export const bandwidthAPI = {
  getProfile: () => api.get('/bandwidth/profile'),
  switchMode: (mode) => api.post('/bandwidth/switch', { mode }),
  reportNetwork: (bandwidthKbps) => api.post('/bandwidth/report', { bandwidth_kbps: bandwidthKbps }),
  getPresets: () => api.get('/bandwidth/presets'),
};

export default api;
