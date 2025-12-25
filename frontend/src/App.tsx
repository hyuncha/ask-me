import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import './styles/App.css';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import AdminKnowledgePage from './pages/AdminKnowledgePage';

// Pages (to be created)
// import ChatPage from './pages/ChatPage';
// import SubscriptionPage from './pages/SubscriptionPage';

function App() {
  return (
    <Router>
      <div className="App">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/chat" element={<div>Chat Page - Coming Soon</div>} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/admin/knowledge" element={<AdminKnowledgePage />} />
          <Route path="/admin" element={<div>Admin Dashboard - Coming Soon</div>} />
          <Route path="/subscription" element={<div>Subscription Page - Coming Soon</div>} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
