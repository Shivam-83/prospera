import React, { useState, useEffect, useRef } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkBreaks from 'remark-breaks';
import { MdSend } from 'react-icons/md';
import Logo from '../Assets/salary_bench.png';
import backgroundImg from '../Assets/women5.png';
import botAvatar from '../Assets/bot.png';
import { Link } from 'react-router-dom';
import backIcon from '../Assets/undo.png';

const ChatSalary = () => {
  const [message, setMessage] = useState('');
  const [chatMessages, setChatMessages] = useState([]);
  const [ws, setWs] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const messagesEndRef = useRef(null);

  const API_BASE = process.env.REACT_APP_API_URL || "https://prospera-bnny.onrender.com";
  const WS_BASE = API_BASE.replace("https://", "wss://").replace("http://", "ws://");

  // Auto-scroll to the latest message.
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatMessages]);

  useEffect(() => {
    const storedUserId = localStorage.getItem("userId");
    if (!storedUserId) {
      alert("User session not found. Please start from the beginning.");
      window.location.href = "/input-form";
      return;
    }

    const wsUrl = `${WS_BASE}/ws/salary?userID=${storedUserId}`;
    console.log("Connecting to WS:", wsUrl);

    const socket = new WebSocket(wsUrl);

    socket.onopen = () => {
      console.log('Salary WebSocket connected');
      setIsConnected(true);
      setIsLoading(false);
    };

    socket.onmessage = (event) => {
      const botMessage = event.data;
      // If the AI session expired, surface the error message clearly.
      setChatMessages((prev) => [...prev, { text: botMessage, sender: 'bot' }]);
    };

    socket.onclose = (event) => {
      console.log('Salary WebSocket disconnected, code:', event.code, 'reason:', event.reason);
      setIsConnected(false);
      setIsLoading(false);
    };

    socket.onerror = (error) => {
      console.error('Salary WebSocket error:', error);
      setIsConnected(false);
      setIsLoading(false);
    };

    setWs(socket);

    return () => {
      socket.close();
    };
  }, []);

  const handleSendMessage = () => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      alert('Chat connection is not open. Please go back and reopen the chat.');
      return;
    }
    if (message.trim()) {
      // Send as plain text — the backend reads it with ws.ReadMessage() directly.
      ws.send(message);

      setChatMessages((prev) => [...prev, { text: message, sender: 'user' }]);
      setMessage('');
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  return (
    <div className="chatbot-page">
      <div className="background-blur" style={{ backgroundImage: `url(${backgroundImg})` }}></div>

      <div className="chat-container">
        <div className="chatbot-header">
          <h2 className="chat-title">Salary Benchmark</h2>
          {/* Connection status indicator */}
          <span style={{
            fontSize: '0.75rem',
            color: isConnected ? '#4caf50' : '#f44336',
            marginLeft: '10px'
          }}>
            {isLoading ? '⏳ Connecting...' : isConnected ? '🟢 Connected' : '🔴 Disconnected'}
          </span>
        </div>

        <div className="chat-messages">
          {isLoading && (
            <div className="chat-message bot-message" style={{ fontStyle: 'italic', opacity: 0.7 }}>
              Connecting to AI coach...
            </div>
          )}
          {chatMessages.map((msg, index) => (
            <div key={index} className={`chat-message ${msg.sender}-message`}>
              {msg.sender === 'bot' ? (
                <>
                  <div className="profile-pic" style={{ backgroundImage: `url(${botAvatar})` }}></div>
                  <ReactMarkdown className="message-text" remarkPlugins={[remarkBreaks]}>
                    {msg.text}
                  </ReactMarkdown>
                </>
              ) : (
                <div className="message-text">{msg.text}</div>
              )}
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>

        <div className="rechoose-option-icon-container">
          <Link to="/chatSuggestions">
            <img src={backIcon} alt="Back" className="rechoose-icon" />
          </Link>
        </div>

        <div className="input-section">
          <input
            type="text"
            className="message-input"
            placeholder={isConnected ? "Type your message..." : "Connecting..."}
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={!isConnected}
          />
          <button className="send-button" onClick={handleSendMessage} disabled={!isConnected}>
            <MdSend className="send-icon" />
          </button>
        </div>
      </div>
    </div>
  );
};

export default ChatSalary;
