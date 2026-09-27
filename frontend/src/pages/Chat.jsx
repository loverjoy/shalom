import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { chatAPI } from '../services/api';
import useStore from '../context/store';

const REACTION_EMOJIS = ['👍', '❤️', '😂', '😮', '😢', '🎉'];

export default function Chat() {
  const { roomId } = useParams();
  const navigate = useNavigate();
  const { user } = useStore();

  const [rooms, setRooms] = useState([]);
  const [activeRoom, setActiveRoom] = useState(roomId || null);
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [showNewRoom, setShowNewRoom] = useState(false);
  const [newRoom, setNewRoom] = useState({ name: '', type: 'group' });
  const [showPoll, setShowPoll] = useState(false);
  const [pollForm, setPollForm] = useState({ question: '', options: ['', ''], is_anonymous: false });
  const [replyTo, setReplyTo] = useState(null);
  const [editingMsg, setEditingMsg] = useState(null);
  const [editContent, setEditContent] = useState('');
  const messagesEnd = useRef(null);

  useEffect(() => { loadRooms(); }, []);
  useEffect(() => { if (activeRoom) loadMessages(activeRoom); }, [activeRoom]);
  useEffect(() => { messagesEnd.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);

  const loadRooms = async () => {
    try {
      const { data } = await chatAPI.getRooms();
      setRooms(data || []);
      if (!activeRoom && data?.length > 0) setActiveRoom(data[0].id);
    } catch (err) {
      console.error('Failed to load rooms');
    }
  };

  const loadMessages = async (roomId) => {
    try {
      const { data } = await chatAPI.getMessages(roomId, { limit: 100 });
      setMessages(data || []);
    } catch (err) {
      console.error('Failed to load messages');
    }
  };

  const handleSendMessage = async (e) => {
    e.preventDefault();
    if (!input.trim() || !activeRoom) return;

    try {
      const payload = {
        chat_room_id: activeRoom,
        content: input,
        reply_to: replyTo?.id || null,
      };
      const { data } = await chatAPI.sendText(payload);
      setMessages((prev) => [...prev, data]);
      setInput('');
      setReplyTo(null);
    } catch (err) {
      toast.error('Failed to send message');
    }
  };

  const handleCreateRoom = async (e) => {
    e.preventDefault();
    try {
      const { data } = await chatAPI.createRoom({
        type: newRoom.type,
        name: newRoom.name,
        members: [user?.id],
      });
      setRooms((prev) => [...prev, data]);
      setActiveRoom(data.id);
      setShowNewRoom(false);
      setNewRoom({ name: '', type: 'group' });
      toast.success('Room created!');
    } catch (err) {
      toast.error('Failed to create room');
    }
  };

  const handleSendPoll = async (e) => {
    e.preventDefault();
    if (!pollForm.question || pollForm.options.filter(Boolean).length < 2) {
      toast.error('Need a question and at least 2 options');
      return;
    }
    try {
      const { data } = await chatAPI.sendPoll({
        chat_room_id: activeRoom,
        question: pollForm.question,
        options: pollForm.options.filter(Boolean),
        is_anonymous: pollForm.is_anonymous,
      });
      setMessages((prev) => [...prev, data]);
      setShowPoll(false);
      setPollForm({ question: '', options: ['', ''], is_anonymous: false });
      toast.success('Poll sent!');
    } catch (err) {
      toast.error('Failed to send poll');
    }
  };

  const handleReaction = async (msgId, emoji) => {
    try {
      await chatAPI.addReaction({ message_id: msgId, emoji });
      setMessages((prev) =>
        prev.map((m) =>
          m.id === msgId
            ? { ...m, reactions: [...(m.reactions || []), { user_id: user.id, emoji }] }
            : m
        )
      );
    } catch (err) {
      console.error('Failed to add reaction');
    }
  };

  const handleEdit = async (msgId) => {
    if (!editContent.trim()) return;
    try {
      await chatAPI.editMessage({ message_id: msgId, content: editContent });
      setMessages((prev) =>
        prev.map((m) => (m.id === msgId ? { ...m, content: editContent, is_edited: true } : m))
      );
      setEditingMsg(null);
      setEditContent('');
    } catch (err) {
      toast.error('Failed to edit');
    }
  };

  const handleDelete = async (msgId) => {
    try {
      await chatAPI.deleteMessage({ message_id: msgId });
      setMessages((prev) => prev.filter((m) => m.id !== msgId));
      toast.success('Message deleted');
    } catch (err) {
      toast.error('Failed to delete');
    }
  };

  return (
    <div className="h-screen flex bg-dark-900">
      {/* Sidebar - Rooms */}
      <div className="w-64 bg-dark-800 border-r border-dark-700 flex flex-col">
        <div className="px-4 py-4 border-b border-dark-700">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-bold text-primary-400">Shalom</h2>
            <button onClick={() => navigate('/dashboard')} className="text-slate-400 hover:text-white text-sm">
              ← Back
            </button>
          </div>
        </div>

        <div className="px-4 py-3 flex items-center justify-between">
          <span className="text-sm text-slate-400">Chats</span>
          <button
            onClick={() => setShowNewRoom(!showNewRoom)}
            className="text-primary-400 hover:text-primary-300 text-xl"
          >
            +
          </button>
        </div>

        {showNewRoom && (
          <form onSubmit={handleCreateRoom} className="px-4 pb-3 space-y-2 animate-fade-in">
            <input
              type="text"
              value={newRoom.name}
              onChange={(e) => setNewRoom({ ...newRoom, name: e.target.value })}
              placeholder="Room name"
              className="w-full bg-dark-900 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-primary-500"
              required
            />
            <select
              value={newRoom.type}
              onChange={(e) => setNewRoom({ ...newRoom, type: e.target.value })}
              className="w-full bg-dark-900 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none"
            >
              <option value="group">Group Chat</option>
              <option value="public">Public Channel</option>
            </select>
            <button type="submit" className="w-full bg-primary-600 hover:bg-primary-700 text-white py-2 rounded-lg text-sm">
              Create
            </button>
          </form>
        )}

        <div className="flex-1 overflow-y-auto">
          {rooms.map((room) => (
            <button
              key={room.id}
              onClick={() => setActiveRoom(room.id)}
              className={`w-full text-left px-4 py-3 border-b border-dark-700 hover:bg-dark-700 transition ${
                activeRoom === room.id ? 'bg-dark-700 border-l-2 border-l-primary-500' : ''
              }`}
            >
              <div className="font-medium text-sm truncate">{room.name}</div>
              <div className="text-xs text-slate-400 mt-0.5">
                {room.type} · {room.message_count || 0} messages
              </div>
            </button>
          ))}
          {rooms.length === 0 && (
            <div className="px-4 py-8 text-center text-slate-500 text-sm">
              No chat rooms yet. Create one!
            </div>
          )}
        </div>
      </div>

      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col">
        {activeRoom ? (
          <>
            {/* Chat Header */}
            <div className="px-6 py-4 border-b border-dark-700 flex items-center justify-between">
              <div>
                <h3 className="font-semibold">
                  {rooms.find((r) => r.id === activeRoom)?.name || 'Chat'}
                </h3>
                <p className="text-xs text-slate-400">
                  {rooms.find((r) => r.id === activeRoom)?.type} room
                </p>
              </div>
              <button
                onClick={() => setShowPoll(!showPoll)}
                className="bg-dark-700 hover:bg-dark-600 px-3 py-2 rounded-lg text-sm transition"
              >
                📊 Poll
              </button>
            </div>

            {/* Poll Form */}
            {showPoll && (
              <div className="px-6 py-4 border-b border-dark-700 bg-dark-800 animate-fade-in">
                <form onSubmit={handleSendPoll} className="space-y-3">
                  <input
                    type="text"
                    value={pollForm.question}
                    onChange={(e) => setPollForm({ ...pollForm, question: e.target.value })}
                    placeholder="Poll question"
                    className="w-full bg-dark-900 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-primary-500"
                    required
                  />
                  {pollForm.options.map((opt, i) => (
                    <input
                      key={i}
                      type="text"
                      value={opt}
                      onChange={(e) => {
                        const opts = [...pollForm.options];
                        opts[i] = e.target.value;
                        setPollForm({ ...pollForm, options: opts });
                      }}
                      placeholder={`Option ${i + 1}`}
                      className="w-full bg-dark-900 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-primary-500"
                    />
                  ))}
                  <button
                    type="button"
                    onClick={() => setPollForm({ ...pollForm, options: [...pollForm.options, ''] })}
                    className="text-primary-400 text-sm hover:text-primary-300"
                  >
                    + Add option
                  </button>
                  <div className="flex items-center gap-4">
                    <label className="flex items-center gap-2 text-sm text-slate-400">
                      <input
                        type="checkbox"
                        checked={pollForm.is_anonymous}
                        onChange={(e) => setPollForm({ ...pollForm, is_anonymous: e.target.checked })}
                      />
                      Anonymous
                    </label>
                    <button type="submit" className="bg-primary-600 hover:bg-primary-700 px-4 py-2 rounded-lg text-sm">
                      Send Poll
                    </button>
                  </div>
                </form>
              </div>
            )}

            {/* Messages */}
            <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
              {messages.map((msg) => (
                <MessageBubble
                  key={msg.id}
                  message={msg}
                  isOwn={msg.sender_id === user?.id}
                  onReply={() => setReplyTo(msg)}
                  onEdit={() => { setEditingMsg(msg.id); setEditContent(msg.content); }}
                  onDelete={() => handleDelete(msg.id)}
                  onReact={(emoji) => handleReaction(msg.id, emoji)}
                />
              ))}
              <div ref={messagesEnd} />
            </div>

            {/* Reply indicator */}
            {replyTo && (
              <div className="px-6 py-2 bg-dark-800 border-t border-dark-700 flex items-center justify-between">
                <div className="text-sm text-slate-400">
                  Replying to <span className="text-white">{replyTo.sender_name}</span>: {replyTo.content}
                </div>
                <button onClick={() => setReplyTo(null)} className="text-slate-400 hover:text-white">✕</button>
              </div>
            )}

            {/* Input */}
            <form onSubmit={handleSendMessage} className="px-6 py-4 border-t border-dark-700 flex gap-3">
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="Type a message..."
                className="flex-1 bg-dark-900 border border-dark-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary-500 transition"
              />
              <button
                type="submit"
                className="bg-primary-600 hover:bg-primary-700 px-6 py-3 rounded-lg font-semibold transition"
              >
                Send
              </button>
            </form>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-slate-500">
            Select or create a chat room
          </div>
        )}
      </div>
    </div>
  );
}

function MessageBubble({ message, isOwn, onReply, onEdit, onDelete, onReact }) {
  const [showActions, setShowActions] = useState(false);
  const [showReactions, setShowReactions] = useState(false);

  if (message.is_deleted) return null;

  return (
    <div
      className={`chat-message group ${isOwn ? 'ml-auto' : ''} max-w-2xl`}
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => { setShowActions(false); setShowReactions(false); }}
    >
      <div className="flex items-start gap-2">
        <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-semibold flex-shrink-0 ${
          isOwn ? 'bg-primary-600' : 'bg-dark-600'
        }`}>
          {message.sender_name?.[0]?.toUpperCase() || '?'}
        </div>
        <div className="flex-1">
          <div className="flex items-baseline gap-2">
            <span className="text-sm font-medium">{message.sender_name}</span>
            <span className="text-xs text-slate-500">
              {new Date(message.created_at).toLocaleTimeString()}
            </span>
            {message.is_edited && <span className="text-xs text-slate-500">(edited)</span>}
          </div>

          {message.type === 'poll' && message.poll ? (
            <PollMessage poll={message.poll} />
          ) : (
            <div className="text-sm text-slate-200 mt-1">{message.content}</div>
          )}

          {message.file_url && (
            <a href={message.file_url} className="text-primary-400 text-sm hover:underline" target="_blank" rel="noreferrer">
              📎 {message.file_name}
            </a>
          )}

          {message.reply_to && (
            <div className="text-xs text-slate-500 mt-1 border-l-2 border-dark-600 pl-2">
              Reply to message
            </div>
          )}

          {/* Reactions */}
          {message.reactions?.length > 0 && (
            <div className="flex gap-1 mt-1">
              {Object.entries(
                message.reactions.reduce((acc, r) => {
                  acc[r.emoji] = (acc[r.emoji] || 0) + 1;
                  return acc;
                }, {})
              ).map(([emoji, count]) => (
                <span key={emoji} className="bg-dark-700 px-1.5 py-0.5 rounded text-xs">
                  {emoji} {count}
                </span>
              ))}
            </div>
          )}
        </div>

        {/* Action buttons */}
        {showActions && (
          <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition">
            <button onClick={onReply} className="p-1 hover:bg-dark-700 rounded text-xs" title="Reply">↩️</button>
            <button onClick={() => setShowReactions(!showReactions)} className="p-1 hover:bg-dark-700 rounded text-xs" title="React">😊</button>
            {isOwn && <button onClick={onEdit} className="p-1 hover:bg-dark-700 rounded text-xs" title="Edit">✏️</button>}
            {isOwn && <button onClick={onDelete} className="p-1 hover:bg-dark-700 rounded text-xs" title="Delete">🗑️</button>}
          </div>
        )}
      </div>

      {/* Reaction picker */}
      {showReactions && (
        <div className="flex gap-1 mt-1 ml-10 animate-fade-in">
          {REACTION_EMOJIS.map((emoji) => (
            <button
              key={emoji}
              onClick={() => { onReact(emoji); setShowReactions(false); }}
              className="hover:bg-dark-700 p-1 rounded text-lg"
            >
              {emoji}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

function PollMessage({ poll }) {
  const totalVotes = poll.options.reduce((sum, opt) => sum + (opt.vote_count || 0), 0);

  return (
    <div className="bg-dark-900 rounded-lg p-3 mt-2 border border-dark-700">
      <div className="text-sm font-medium mb-2">📊 {poll.question}</div>
      {poll.options.map((opt, i) => {
        const pct = totalVotes > 0 ? Math.round((opt.vote_count / totalVotes) * 100) : 0;
        return (
          <div key={i} className="mb-2">
            <div className="flex justify-between text-xs text-slate-400 mb-0.5">
              <span>{opt.text}</span>
              <span>{pct}% ({opt.vote_count})</span>
            </div>
            <div className="w-full bg-dark-700 rounded-full h-2">
              <div className="bg-primary-500 h-2 rounded-full transition-all" style={{ width: `${pct}%` }} />
            </div>
          </div>
        );
      })}
      <div className="text-xs text-slate-500 mt-2">{totalVotes} vote{totalVotes !== 1 ? 's' : ''}</div>
    </div>
  );
}
