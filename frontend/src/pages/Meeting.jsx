import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Room, RoomEvent, Track, ConnectionState, Participant } from 'livekit-client';
import toast from 'react-hot-toast';
import { meetingAPI, bandwidthAPI } from '../services/api';
import useStore from '../context/store';

const BANDWIDTH_MODES = {
  ultra_saving: { label: 'Ultra Saving', desc: 'Audio Only (8-15 MB/hr)', icon: '🎧' },
  economy: { label: 'Economy', desc: '144p Video (50-100 MB/hr)', icon: '📉' },
  standard: { label: 'Standard', desc: '360p Video (150-250 MB/hr)', icon: '📺' },
  hd: { label: 'HD', desc: '720p Video (500-800 MB/hr)', icon: '🖥️' },
};

export default function Meeting() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user, localMuted, localVideoOn, localScreenShare, handRaised, bandwidthMode,
          toggleMute, toggleVideo, toggleScreenShare, toggleHandRaise, setBandwidthMode } = useStore();

  const [meeting, setMeeting] = useState(null);
  const [room, setRoom] = useState(null);
  const [participants, setParticipants] = useState([]);
  const [connected, setConnected] = useState(false);
  const [showChat, setShowChat] = useState(false);
  const [showParticipants, setShowParticipants] = useState(false);
  const [networkKbps, setNetworkKbps] = useState(500);

  const videoRef = useRef(null);
  const roomRef = useRef(null);

  useEffect(() => {
    joinMeeting();
    return () => { if (roomRef.current) roomRef.current.disconnect(); };
  }, []);

  useEffect(() => {
    const interval = setInterval(detectNetwork, 5000);
    return () => clearInterval(interval);
  }, []);

  const joinMeeting = async () => {
    try {
      // Get meeting info
      const { data: meetingData } = await meetingAPI.get(id);
      setMeeting(meetingData);

      // Join via API to get LiveKit token
      const { data } = await meetingAPI.join({ meeting_code: meetingData.meeting_code });

      // Connect to LiveKit
      const lkRoom = new Room({
        adaptiveStream: true,
        dynacast: true,
        audioCaptureDefaults: { echoCancellation: true, noiseSuppression: true },
        videoCaptureDefaults: { resolution: { width: 1280, height: 720 } },
      });

      lkRoom.on(RoomEvent.Connected, () => {
        setConnected(true);
        toast.success('Connected to meeting');
      });

      lkRoom.on(RoomEvent.Disconnected, () => {
        toast.error('Disconnected from meeting');
        setConnected(false);
      });

      lkRoom.on(RoomEvent.ParticipantConnected, (participant) => {
        setParticipants(Array.from(lkRoom.participants.values()));
      });

      lkRoom.on(RoomEvent.ParticipantDisconnected, (participant) => {
        setParticipants(Array.from(lkRoom.participants.values()));
      });

      lkRoom.on(RoomEvent.TrackSubscribed, (track, publication, participant) => {
        if (track.kind === Track.Kind.Video) {
          const el = document.getElementById(`video-${participant.identity}`);
          if (el) track.attach(el);
        }
        if (track.kind === Track.Kind.Audio) {
          const el = document.getElementById(`audio-${participant.identity}`);
          if (el) track.attach(el);
        }
      });

      await lkRoom.connect(data.room_name, data.token);
      roomRef.current = lkRoom;
      setRoom(lkRoom);
      setParticipants(Array.from(lkRoom.participants.values()));

      // Publish local tracks
      await lkRoom.localParticipant.setMicrophoneEnabled(!localMuted);
      await lkRoom.localParticipant.setCameraEnabled(localVideoOn);

    } catch (err) {
      toast.error('Failed to join meeting');
      console.error(err);
      navigate('/dashboard');
    }
  };

  const detectNetwork = async () => {
    if (!navigator.connection) return;
    const conn = navigator.connection;
    const kbps = conn.downlink * 1000 || 500;
    setNetworkKbps(kbps);
    try { await bandwidthAPI.reportNetwork(kbps); } catch (e) {}
  };

  const handleToggleMute = async () => {
    toggleMute();
    if (room) await room.localParticipant.setMicrophoneEnabled(!localMuted);
  };

  const handleToggleVideo = async () => {
    toggleVideo();
    if (room) await room.localParticipant.setCameraEnabled(!localVideoOn);
  };

  const handleToggleScreenShare = async () => {
    toggleScreenShare();
    if (room) {
      if (!localScreenShare) {
        await room.localParticipant.setScreenShareEnabled(true);
      } else {
        await room.localParticipant.setScreenShareEnabled(false);
      }
    }
  };

  const handleSwitchMode = async (mode) => {
    setBandwidthMode(mode);
    try { await bandwidthAPI.switchMode(mode); } catch (e) {}

    // Adjust video quality based on mode
    if (room) {
      const preset = BANDWIDTH_MODES[mode];
      if (mode === 'ultra_saving') {
        await room.localParticipant.setCameraEnabled(false);
      } else if (!localVideoOn) {
        await room.localParticipant.setCameraEnabled(true);
      }
    }
  };

  const handleEndMeeting = async () => {
    try {
      await meetingAPI.end(id);
      if (room) room.disconnect();
      navigate('/dashboard');
    } catch (err) {
      toast.error(err.response?.data?.error || 'Failed to end meeting');
    }
  };

  return (
    <div className="h-screen flex flex-col bg-dark-900">
      {/* Top Bar */}
      <div className="bg-dark-800 border-b border-dark-700 px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h2 className="font-semibold">{meeting?.title || 'Meeting'}</h2>
          {connected && <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />}
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-slate-400">{participants.length} participants</span>
          <span className="text-sm text-slate-400">{networkKbps} kbps</span>
          <button onClick={() => navigate('/dashboard')} className="text-slate-400 hover:text-white text-sm">
            Leave
          </button>
        </div>
      </div>

      {/* Video Grid */}
      <div className="flex-1 p-4 overflow-auto">
        <div className="meeting-grid h-full">
          {/* Local video */}
          <div className="bg-dark-800 rounded-xl relative overflow-hidden border border-dark-700 aspect-video">
            <video
              ref={videoRef}
              id={`video-${user?.id}`}
              autoPlay
              muted
              playsInline
              className="w-full h-full object-cover"
            />
            <div className="absolute bottom-2 left-2 bg-black/60 px-2 py-1 rounded text-xs">
              You {localMuted && '(Muted)'}
            </div>
            {handRaised && (
              <div className="absolute top-2 right-2 text-2xl animate-bounce">✋</div>
            )}
          </div>

          {/* Remote participants */}
          {participants.map((p) => (
            <ParticipantTile key={p.identity} participant={p} />
          ))}
        </div>
      </div>

      {/* Bandwidth Mode Selector */}
      <div className="bg-dark-800 border-t border-dark-700 px-6 py-2 flex items-center justify-center gap-2">
        {Object.entries(BANDWIDTH_MODES).map(([mode, info]) => (
          <button
            key={mode}
            onClick={() => handleSwitchMode(mode)}
            className={`px-3 py-1.5 rounded-lg text-xs transition ${
              bandwidthMode === mode
                ? 'bg-primary-600 text-white'
                : 'bg-dark-700 text-slate-400 hover:bg-dark-600'
            }`}
            title={info.desc}
          >
            {info.icon} {info.label}
          </button>
        ))}
      </div>

      {/* Bottom Controls */}
      <div className="bg-dark-800 border-t border-dark-700 px-6 py-4 flex items-center justify-center gap-4">
        <button
          onClick={handleToggleMute}
          className={`p-4 rounded-full transition ${
            localMuted ? 'bg-red-600 hover:bg-red-700' : 'bg-dark-700 hover:bg-dark-600'
          }`}
          title={localMuted ? 'Unmute' : 'Mute'}
        >
          {localMuted ? '🔇' : '🎤'}
        </button>

        <button
          onClick={handleToggleVideo}
          className={`p-4 rounded-full transition ${
            !localVideoOn ? 'bg-red-600 hover:bg-red-700' : 'bg-dark-700 hover:bg-dark-600'
          }`}
          title={localVideoOn ? 'Turn off camera' : 'Turn on camera'}
        >
          {localVideoOn ? '📹' : '📷'}
        </button>

        <button
          onClick={handleToggleScreenShare}
          className={`p-4 rounded-full transition ${
            localScreenShare ? 'bg-primary-600 hover:bg-primary-700' : 'bg-dark-700 hover:bg-dark-600'
          }`}
          title="Share screen"
        >
          🖥️
        </button>

        <button
          onClick={toggleHandRaise}
          className={`p-4 rounded-full transition ${
            handRaised ? 'bg-yellow-600 hover:bg-yellow-700' : 'bg-dark-700 hover:bg-dark-600'
          }`}
          title="Raise hand"
        >
          ✋
        </button>

        <button
          onClick={() => setShowParticipants(!showParticipants)}
          className="p-4 rounded-full bg-dark-700 hover:bg-dark-600 transition"
          title="Participants"
        >
          👥 ({participants.length})
        </button>

        <button
          onClick={() => setShowChat(!showChat)}
          className="p-4 rounded-full bg-dark-700 hover:bg-dark-600 transition"
          title="Chat"
        >
          💬
        </button>

        <button
          onClick={handleEndMeeting}
          className="p-4 rounded-full bg-red-600 hover:bg-red-700 transition ml-4"
          title="End meeting"
        >
          📞
        </button>
      </div>

      {/* Side Chat Panel */}
      {showChat && (
        <MeetingChat meetingCode={meeting?.meeting_code} onClose={() => setShowChat(false)} />
      )}

      {/* Participants Panel */}
      {showParticipants && (
        <ParticipantsPanel participants={participants} onClose={() => setShowParticipants(false)} />
      )}

      {/* Hidden audio elements */}
      {participants.map((p) => (
        <audio key={p.identity} id={`audio-${p.identity}`} autoPlay />
      ))}
    </div>
  );
}

function ParticipantTile({ participant }) {
  const videoRef = useRef(null);

  useEffect(() => {
    if (!videoRef.current) return;
    const videoPub = participant.getTrackPublication(Track.Source.Camera);
    if (videoPub?.track) {
      videoPub.track.attach(videoRef.current);
    }
    return () => {
      if (videoPub?.track) videoPub.track.detach();
    };
  }, [participant]);

  return (
    <div className="bg-dark-800 rounded-xl relative overflow-hidden border border-dark-700 aspect-video">
      <video ref={videoRef} autoPlay playsInline className="w-full h-full object-cover" />
      <div className="absolute bottom-2 left-2 bg-black/60 px-2 py-1 rounded text-xs">
        {participant.identity}
        {participant.isMuted && ' (Muted)'}
      </div>
      {participant.isSpeaking && (
        <div className="absolute inset-0 border-2 border-green-500 rounded-xl pointer-events-none" />
      )}
    </div>
  );
}

function MeetingChat({ meetingCode, onClose }) {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const { user } = useStore();
  const messagesEnd = useRef(null);

  useEffect(() => {
    messagesEnd.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async (e) => {
    e.preventDefault();
    if (!input.trim()) return;
    const msg = {
      id: Date.now(),
      sender: user?.display_name || 'You',
      content: input,
      time: new Date().toLocaleTimeString(),
    };
    setMessages((prev) => [...prev, msg]);
    setInput('');
  };

  return (
    <div className="fixed right-0 top-0 h-full w-80 bg-dark-800 border-l border-dark-700 flex flex-col z-50">
      <div className="px-4 py-3 border-b border-dark-700 flex items-center justify-between">
        <h3 className="font-semibold">Chat</h3>
        <button onClick={onClose} className="text-slate-400 hover:text-white">✕</button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.map((msg) => (
          <div key={msg.id} className="chat-message">
            <div className="text-xs text-slate-400">{msg.sender} · {msg.time}</div>
            <div className="text-sm mt-1">{msg.content}</div>
          </div>
        ))}
        <div ref={messagesEnd} />
      </div>

      <form onSubmit={handleSend} className="p-3 border-t border-dark-700 flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Type a message..."
          className="flex-1 bg-dark-900 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-primary-500"
        />
        <button type="submit" className="bg-primary-600 hover:bg-primary-700 px-3 py-2 rounded-lg text-sm">
          Send
        </button>
      </form>
    </div>
  );
}

function ParticipantsPanel({ participants, onClose }) {
  return (
    <div className="fixed right-0 top-0 h-full w-80 bg-dark-800 border-l border-dark-700 flex flex-col z-50">
      <div className="px-4 py-3 border-b border-dark-700 flex items-center justify-between">
        <h3 className="font-semibold">Participants ({participants.length})</h3>
        <button onClick={onClose} className="text-slate-400 hover:text-white">✕</button>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {participants.map((p) => (
          <div key={p.identity} className="flex items-center gap-3 p-2 rounded-lg hover:bg-dark-700">
            <div className="w-8 h-8 bg-primary-600 rounded-full flex items-center justify-center text-sm font-semibold">
              {p.identity[0]?.toUpperCase()}
            </div>
            <div>
              <div className="text-sm">{p.identity}</div>
              <div className="text-xs text-slate-400">
                {p.isMuted ? '🔇 Muted' : '🎤 Speaking'} · {p.isCameraOn ? '📹' : '📷'}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
