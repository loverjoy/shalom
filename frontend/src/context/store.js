import { create } from 'zustand';

const useStore = create((set, get) => ({
  // Auth
  user: JSON.parse(localStorage.getItem('sm_user') || 'null'),
  token: localStorage.getItem('sm_token') || null,
  isAuthenticated: !!localStorage.getItem('sm_token'),

  setAuth: (user, token) => {
    localStorage.setItem('sm_user', JSON.stringify(user));
    localStorage.setItem('sm_token', token);
    set({ user, token, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem('sm_user');
    localStorage.removeItem('sm_token');
    set({ user: null, token: null, isAuthenticated: false });
  },

  // Meeting
  currentMeeting: null,
  participants: [],
  isHost: false,
  localMuted: false,
  localVideoOn: false,
  localScreenShare: false,
  handRaised: false,
  bandwidthMode: 'standard',

  setCurrentMeeting: (meeting) => set({ currentMeeting: meeting }),
  setParticipants: (participants) => set({ participants }),
  setIsHost: (isHost) => set({ isHost }),
  toggleMute: () => set((s) => ({ localMuted: !s.localMuted })),
  toggleVideo: () => set((s) => ({ localVideoOn: !s.localVideoOn })),
  toggleScreenShare: () => set((s) => ({ localScreenShare: !s.localScreenShare })),
  toggleHandRaise: () => set((s) => ({ handRaised: !s.handRaised })),
  setBandwidthMode: (mode) => set({ bandwidthMode: mode }),

  // Chat
  messages: [],
  chatRooms: [],
  currentChatRoom: null,

  setMessages: (messages) => set({ messages }),
  addMessage: (msg) => set((s) => ({ messages: [...s.messages, msg] })),
  setChatRooms: (rooms) => set({ chatRooms: rooms }),
  setCurrentChatRoom: (room) => set({ currentChatRoom: room }),
}));

export default useStore;
