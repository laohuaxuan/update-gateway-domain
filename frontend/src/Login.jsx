import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

export default function Login({ onLogin }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [captchaCode, setCaptchaCode] = useState("");
  const [captchaToken, setCaptchaToken] = useState("");
  const [captchaImageUrl, setCaptchaImageUrl] = useState("");
  const [message, setMessage] = useState("");
  const [messageType, setMessageType] = useState("success");
  const [loading, setLoading] = useState(false);
  const messageTimeoutRef = useRef(null);
  const navigate = useNavigate();

  const showMessage = (text, type = "success") => {
    if (messageTimeoutRef.current) {
      clearTimeout(messageTimeoutRef.current);
    }
    setMessage(text);
    setMessageType(type);
    messageTimeoutRef.current = setTimeout(() => {
      setMessage("");
    }, 3000);
  };

  const refreshCaptcha = () => {
    const token = crypto.randomUUID().replace(/-/g, "");
    setCaptchaToken(token);
    setCaptchaCode("");
    setCaptchaImageUrl(`/api/captcha?token=${encodeURIComponent(token)}&t=${Date.now()}`);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    showMessage("", "success");

    if (!captchaToken || !captchaCode) {
      showMessage("请输入验证码", "error");
      return;
    }
    if (!/^\d{4}$/.test(captchaCode)) {
      showMessage("验证码必须为4位数字", "error");
      return;
    }

    setLoading(true);

    try {
      const resp = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          username, 
          password,
          captcha_token: captchaToken,
          captcha_code: captchaCode,
        }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "登录失败");
      }
      localStorage.setItem("token", data.token);
      localStorage.setItem("role", data.role);
      localStorage.setItem("name", data.name);
      onLogin && onLogin(data);
      navigate("/");
    } catch (err) {
      showMessage(err.message, "error");
      refreshCaptcha();
    } finally {
      setLoading(false);
    }
  };

  const handleResetPassword = async () => {
    if (!username || !username.trim()) {
      showMessage("请先输入用户名或手机号", "error");
      return;
    }
    try {
      const resp = await fetch("/api/reset-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ account: username.trim() }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "发送重置邮件失败");
      }
      showMessage(data.message || "重置链接已发送到该账号绑定邮箱，请检查邮箱。");
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  useEffect(() => {
    refreshCaptcha();
  }, []);

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h1>欢迎登录MSE网关管理系统</h1>
        <h2>登录</h2>
        {message && <div className={`message ${messageType === 'error' ? 'message-error' : ''}`}>{message}</div>}
        
        <form onSubmit={handleSubmit}>
          <div>
            <label>用户名/手机号</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              placeholder="请输入用户名或手机号"
            />
          </div>
          <div>
            <label>密码</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="请输入密码"
            />
          </div>
          <div className="captcha-section">
            <label>验证码</label>
            <div className="captcha-row">
              <input
                type="text"
                value={captchaCode}
                onChange={(e) => setCaptchaCode(e.target.value.replace(/\D/g, "").slice(0, 4))}
                required
                placeholder="请输入4位数字验证码"
                className="captcha-input"
                inputMode="numeric"
                maxLength={4}
              />
              <img
                src={captchaImageUrl}
                alt="验证码"
                className="captcha-img"
                onClick={refreshCaptcha}
                title="点击刷新验证码"
              />
            </div>
          </div>
          <button type="submit" disabled={loading}>
            {loading ? "登录中..." : "登录"}
          </button>
        </form>
        <div className="auth-links">
          <button type="button" className="link-btn" onClick={handleResetPassword}>
            忘记密码？
          </button>
          <span>或</span>
          <Link to="/register">注册新账号</Link>
        </div>
      </div>
    </div>
  );
}
