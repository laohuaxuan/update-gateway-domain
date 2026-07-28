import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

function EyeToggleIcon({ visible }) {
  if (visible) {
    return (
      <svg className="eye-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path
          d="M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6-10-6-10-6z"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="1.8" />
      </svg>
    );
  }
  return (
    <svg className="eye-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path
        d="M3 3l18 18"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M10.6 6.5A11.6 11.6 0 0 1 12 6c6.5 0 10 6 10 6a17 17 0 0 1-3 3.7M6.2 8.2A17 17 0 0 0 2 12s3.5 6 10 6c1.2 0 2.3-.2 3.3-.6"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function LdapIcon() {
  return (
    <svg className="third-party-icon" viewBox="0 0 24 24" aria-hidden="true">
      <path
        fill="currentColor"
        d="M12 2a4 4 0 0 1 4 4v1h1.5A2.5 2.5 0 0 1 20 9.5V11h-2v-.5a.5.5 0 0 0-.5-.5H14V6a2 2 0 1 0-4 0v4H6.5a.5.5 0 0 0-.5.5V11H4V9.5A2.5 2.5 0 0 1 6.5 7H8V6a4 4 0 0 1 4-4zm-6 11h12v2H6v-2zm1 4h4v2H7v-2zm6 0h4v2h-4v-2z"
      />
    </svg>
  );
}

function persistLogin(data) {
  localStorage.setItem("token", data.token);
  localStorage.setItem("role", data.role || "");
  localStorage.setItem("name", data.name || "");
  localStorage.setItem("email", data.email || "");
  localStorage.setItem("phone", data.phone || "");
  localStorage.setItem("auth_source", data.auth_source || "local");
}

export default function Login({ onLogin }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [captchaCode, setCaptchaCode] = useState("");
  const [captchaToken, setCaptchaToken] = useState("");
  const [captchaImageUrl, setCaptchaImageUrl] = useState("");
  const [message, setMessage] = useState("");
  const [messageType, setMessageType] = useState("success");
  const [loading, setLoading] = useState(false);
  const [ldapLabel, setLdapLabel] = useState("LDAP");
  const [ldapEnabled, setLdapEnabled] = useState(false);
  const [ldapModalOpen, setLdapModalOpen] = useState(false);
  const [ldapUsername, setLdapUsername] = useState("");
  const [ldapPassword, setLdapPassword] = useState("");
  const [ldapPasswordVisible, setLdapPasswordVisible] = useState(false);
  const [ldapLoading, setLdapLoading] = useState(false);
  const [ldapMessage, setLdapMessage] = useState("");
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

  const showLdapMessage = (text) => {
    setLdapMessage(text || "");
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
      persistLogin(data);
      onLogin && onLogin(data);
      navigate("/");
    } catch (err) {
      showMessage(err.message, "error");
      refreshCaptcha();
    } finally {
      setLoading(false);
    }
  };

  const submitLdapLogin = async (e) => {
    e.preventDefault();
    showLdapMessage("");
    if (!ldapUsername.trim() || !ldapPassword) {
      showLdapMessage("请输入 LDAP 用户名和密码");
      return;
    }
    setLdapLoading(true);
    try {
      const resp = await fetch("/api/login/ldap", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          username: ldapUsername.trim(),
          password: ldapPassword,
        }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "LDAP 登录失败");
      }
      persistLogin(data);
      onLogin && onLogin(data);
      setLdapModalOpen(false);
      showLdapMessage("");
      navigate("/");
    } catch (err) {
      showLdapMessage(err.message || "LDAP 登录失败");
    } finally {
      setLdapLoading(false);
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
    fetch("/api/auth/providers")
      .then((r) => r.json())
      .then((data) => {
        const providers = Array.isArray(data.providers) ? data.providers : [];
        const ldap = providers.find((p) => p.id === "ldap" || p.type === "ldap");
        if (ldap) {
          setLdapEnabled(true);
          setLdapLabel(ldap.label || "LDAP");
        }
      })
      .catch(() => {});
  }, []);

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h1>欢迎登录MSE网关管理系统</h1>
        <h2>登录</h2>
        {message && <div className={`message ${messageType === "error" ? "message-error" : ""}`}>{message}</div>}

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
            <div className="password-input-row">
              <input
                type={passwordVisible ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                placeholder="请输入密码"
              />
              <button
                type="button"
                className="password-toggle-btn"
                title={passwordVisible ? "隐藏密码" : "显示密码"}
                aria-label={passwordVisible ? "隐藏密码" : "显示密码"}
                onClick={() => setPasswordVisible((v) => !v)}
              >
                <EyeToggleIcon visible={passwordVisible} />
              </button>
            </div>
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
        {ldapEnabled && (
          <div className="third-party-login">
            <span className="third-party-label">第三方登录：</span>
            <button
              type="button"
              className="third-party-btn"
              title={`${ldapLabel} 登录`}
              aria-label={`${ldapLabel} 登录`}
              onClick={() => {
                setLdapUsername("");
                setLdapPassword("");
                setLdapPasswordVisible(false);
                setLdapMessage("");
                setLdapModalOpen(true);
              }}
            >
              <LdapIcon />
              <span>{ldapLabel}</span>
            </button>
          </div>
        )}
        <div className="auth-links">
          <button type="button" className="link-btn" onClick={handleResetPassword}>
            忘记密码？
          </button>
          <span>或</span>
          <Link to="/register">注册新账号</Link>
        </div>
      </div>

      {ldapModalOpen && (
        <div
          className="modal-mask"
          onClick={() => {
            if (!ldapLoading) {
              setLdapModalOpen(false);
              setLdapMessage("");
            }
          }}
        >
          <div className="modal-card ldap-login-modal" onClick={(e) => e.stopPropagation()}>
            <h2>{ldapLabel} 登录</h2>
            {ldapMessage && <div className="message message-error">{ldapMessage}</div>}
            <form onSubmit={submitLdapLogin}>
              <div className="form-group">
                <label>用户名</label>
                <input
                  type="text"
                  value={ldapUsername}
                  onChange={(e) => setLdapUsername(e.target.value)}
                  required
                  autoFocus
                  placeholder="请输入 LDAP 用户名"
                  autoComplete="username"
                />
              </div>
              <div className="form-group">
                <label>密码</label>
                <div className="password-input-row">
                  <input
                    type={ldapPasswordVisible ? "text" : "password"}
                    value={ldapPassword}
                    onChange={(e) => setLdapPassword(e.target.value)}
                    required
                    placeholder="请输入 LDAP 密码"
                    autoComplete="current-password"
                  />
                  <button
                    type="button"
                    className="password-toggle-btn"
                    title={ldapPasswordVisible ? "隐藏密码" : "显示密码"}
                    aria-label={ldapPasswordVisible ? "隐藏密码" : "显示密码"}
                    onClick={() => setLdapPasswordVisible((v) => !v)}
                  >
                    <EyeToggleIcon visible={ldapPasswordVisible} />
                  </button>
                </div>
              </div>
              <div className="row">
                <button
                  type="button"
                  disabled={ldapLoading}
                  onClick={() => {
                    setLdapModalOpen(false);
                    setLdapMessage("");
                  }}
                >
                  取消
                </button>
                <button type="submit" disabled={ldapLoading}>
                  {ldapLoading ? "登录中..." : "登录"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
