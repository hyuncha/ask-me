import React from 'react';
import '../styles/LoginPage.css';

const LoginPage: React.FC = () => {
  const handleGoogleLogin = () => {
    // Redirect to backend Google OAuth
    window.location.href = 'http://localhost:8080/auth/google';
  };

  return (
    <div className="login-page">
      <div className="login-container">
        <div className="login-header">
          <h1>Cleaners AI</h1>
          <p>AI 기반 세탁 지식 서비스</p>
        </div>

        <div className="login-content">
          <h2>로그인</h2>
          <p className="login-subtitle">
            세탁에 관한 모든 것을 AI와 함께 알아보세요
          </p>

          <button className="google-login-button" onClick={handleGoogleLogin}>
            <svg className="google-icon" viewBox="0 0 24 24">
              <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
            </svg>
            구글로 로그인
          </button>

          <div className="login-features">
            <div className="feature">
              <span className="feature-icon">💬</span>
              <span>AI 채팅 상담</span>
            </div>
            <div className="feature">
              <span className="feature-icon">📚</span>
              <span>세탁 지식 베이스</span>
            </div>
            <div className="feature">
              <span className="feature-icon">📊</span>
              <span>대화 기록 저장</span>
            </div>
          </div>
        </div>

        <div className="login-footer">
          <p>로그인하시면 <a href="/terms">이용약관</a> 및 <a href="/privacy">개인정보처리방침</a>에 동의하게 됩니다.</p>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
