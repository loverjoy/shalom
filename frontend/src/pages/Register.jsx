import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { authAPI } from '../services/api';
import useStore from '../context/store';

export default function Register() {
  const [form, setForm] = useState({ email: '', username: '', display_name: '', password: '' });
  const [loading, setLoading] = useState(false);
  const setAuth = useStore((s) => s.setAuth);
  const navigate = useNavigate();

  const handleChange = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (form.password.length < 8) {
      toast.error('Password must be at least 8 characters');
      return;
    }
    setLoading(true);
    try {
      const { data } = await authAPI.register(form);
      setAuth(data.user, data.token);
      toast.success('Account created!');
      navigate('/dashboard');
    } catch (err) {
      toast.error(err.response?.data?.error || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-primary-400">Shalom</h1>
          <p className="text-slate-400 mt-2">Create your account</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-dark-800 rounded-2xl p-8 shadow-xl border border-dark-700">
          <h2 className="text-2xl font-semibold mb-6">Sign Up</h2>

          <div className="space-y-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Display Name</label>
              <input
                type="text"
                name="display_name"
                value={form.display_name}
                onChange={handleChange}
                className="w-full bg-dark-900 border border-dark-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary-500 transition"
                placeholder="John Doe"
                required
              />
            </div>

            <div>
              <label className="block text-sm text-slate-400 mb-1">Username</label>
              <input
                type="text"
                name="username"
                value={form.username}
                onChange={handleChange}
                className="w-full bg-dark-900 border border-dark-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary-500 transition"
                placeholder="johndoe"
                required
              />
            </div>

            <div>
              <label className="block text-sm text-slate-400 mb-1">Email</label>
              <input
                type="email"
                name="email"
                value={form.email}
                onChange={handleChange}
                className="w-full bg-dark-900 border border-dark-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary-500 transition"
                placeholder="you@example.com"
                required
              />
            </div>

            <div>
              <label className="block text-sm text-slate-400 mb-1">Password</label>
              <input
                type="password"
                name="password"
                value={form.password}
                onChange={handleChange}
                className="w-full bg-dark-900 border border-dark-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary-500 transition"
                placeholder="Min 8 characters"
                required
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full mt-6 bg-primary-600 hover:bg-primary-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition"
          >
            {loading ? 'Creating account...' : 'Create Account'}
          </button>

          <p className="text-center text-slate-400 mt-6">
            Already have an account?{' '}
            <Link to="/login" className="text-primary-400 hover:text-primary-300">Sign In</Link>
          </p>
        </form>
      </div>
    </div>
  );
}
