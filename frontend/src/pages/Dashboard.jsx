import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { meetingAPI, authAPI } from '../services/api';
import useStore from '../context/store';

export default function Dashboard() {
  const [meetings, setMeetings] = useState([]);
  const [joinCode, setJoinCode] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [createForm, setCreateForm] = useState({ title: '', description: '' });
  const [loading, setLoading] = useState(false);
  const { user, logout, setAuth } = useStore();
  const navigate = useNavigate();

  useEffect(() => {
    loadProfile();
  }, []);

  const loadProfile = async () => {
    try {
      const { data } = await authAPI.getProfile();
      setAuth(data, localStorage.getItem('sm_token'));
    } catch (err) {
      console.error('Failed to load profile');
    }
  };

  const handleCreateMeeting = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      const { data } = await meetingAPI.create(createForm);
      toast.success('Meeting created!');
      setShowCreate(false);
      setCreateForm({ title: '', description: '' });
      navigate(`/meeting/${data.id}`);
    } catch (err) {
      toast.error(err.response?.data?.error || 'Failed to create meeting');
    } finally {
      setLoading(false);
    }
  };

  const handleJoinMeeting = async (e) => {
    e.preventDefault();
    if (!joinCode.trim()) return;
    setLoading(true);
    try {
      const { data } = await meetingAPI.join({ meeting_code: joinCode.trim() });
      toast.success('Joined meeting!');
      navigate(`/meeting/${data.meeting.id}`);
    } catch (err) {
      toast.error(err.response?.data?.error || 'Failed to join meeting');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try { await authAPI.logout(); } catch (e) {}
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-dark-900">
      {/* Header */}
      <header className="bg-dark-800 border-b border-dark-700 px-6 py-4">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <h1 className="text-2xl font-bold text-primary-400">Shalom</h1>
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate('/chat')}
              className="text-slate-300 hover:text-white transition px-3 py-2 rounded-lg hover:bg-dark-700"
            >
              Chat
            </button>
            <span className="text-slate-400">{user?.display_name}</span>
            <button onClick={handleLogout} className="text-slate-400 hover:text-red-400 transition">
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Main */}
      <main className="max-w-6xl mx-auto px-6 py-10">
        {/* Quick Join */}
        <div className="bg-dark-800 rounded-2xl p-8 border border-dark-700 mb-8">
          <h2 className="text-xl font-semibold mb-4">Join a Meeting</h2>
          <form onSubmit={handleJoinMeeting} className="flex gap-3">
            <input
              type="text"
              value={joinCode}
              onChange={(e) => setJoinCode(e.target.value)}
              placeholder="Enter meeting code"
              className="flex-1 bg-dark-900 border border-dark-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary-500 transition"
            />
            <button
              type="submit"
              disabled={loading}
              className="bg-primary-600 hover:bg-primary-700 disabled:opacity-50 text-white font-semibold px-6 py-3 rounded-lg transition"
            >
              Join
            </button>
          </form>
        </div>

        {/* Create Meeting */}
        <div className="bg-dark-800 rounded-2xl p-8 border border-dark-700 mb-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Start a New Meeting</h2>
            <button
              onClick={() => setShowCreate(!showCreate)}
              className="bg-green-600 hover:bg-green-700 text-white font-semibold px-6 py-3 rounded-lg transition"
            >
              {showCreate ? 'Cancel' : 'New Meeting'}
            </button>
          </div>

          {showCreate && (
            <form onSubmit={handleCreateMeeting} className="space-y-4 animate-fade-in">
              <div>
                <label className="block text-sm text-slate-400 mb-1">Meeting Title</label>
                <input
                  type="text"
                  value={createForm.title}
                  onChange={(e) => setCreateForm({ ...createForm, title: e.target.value })}
                  className="w-full bg-dark-900 border border-dark-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary-500 transition"
                  placeholder="Team Standup"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Description (optional)</label>
                <input
                  type="text"
                  value={createForm.description}
                  onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
                  className="w-full bg-dark-900 border border-dark-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary-500 transition"
                  placeholder="Daily sync"
                />
              </div>
              <button
                type="submit"
                disabled={loading}
                className="w-full bg-primary-600 hover:bg-primary-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition"
              >
                {loading ? 'Creating...' : 'Create & Start Meeting'}
              </button>
            </form>
          )}
        </div>

        {/* Info Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-dark-800 rounded-2xl p-6 border border-dark-700">
            <div className="text-3xl mb-3">500+</div>
            <div className="text-slate-400">Participants per meeting</div>
          </div>
          <div className="bg-dark-800 rounded-2xl p-6 border border-dark-700">
            <div className="text-3xl mb-3">70-90%</div>
            <div className="text-slate-400">Less data usage</div>
          </div>
          <div className="bg-dark-800 rounded-2xl p-6 border border-dark-700">
            <div className="text-3xl mb-3">2G-5G</div>
            <div className="text-slate-400">Works on any network</div>
          </div>
        </div>
      </main>
    </div>
  );
}
