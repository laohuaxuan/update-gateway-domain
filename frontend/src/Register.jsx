import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

const hasUppercase = /[A-Z]/;
const hasLowercase = /[a-z]/;
const hasDigit = /\d/;

const isPasswordValid = (password) =>
  password.length >= 6 && hasUppercase.test(password) && hasLowercase.test(password) && hasDigit.test(password);

export default function Register() {
  const [username, setUsername] = useState("");
  const [usernameValid, setUsernameValid] = useState(true);
  const [usernameMessage, setUsernameMessage] = useState("");
  const [phone, setPhone] = useState("");
  const [phoneValid, setPhoneValid] = useState(true);
  const [phoneMessage, setPhoneMessage] = useState("");
  const [email, setEmail] = useState("");
  const [emailValid, setEmailValid] = useState(true);
  const [emailMessage, setEmailMessage] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
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

  const checkExists = async (field, value) => {
    if (!value) return false;
    try {
      const resp = await fetch(`/api/check-exists?field=${field}&value=${encodeURIComponent(value)}`);
      const data = await resp.json();
      return data.exists;
    } catch {
      return false;
    }
  };

  const validateUsername = async (e) => {
    const value = e.target.value;
    setUsername(value);
    
    if (value === "") {
      setUsernameValid(true);
      setUsernameMessage("");
      return;
    }
    
    if (value.length < 3) {
      setUsernameValid(false);
      setUsernameMessage("用户名最少3个字符");
      return;
    }
    
    if (value.length > 24) {
      setUsernameValid(false);
      setUsernameMessage("用户名最长不超过24个字符");
      return;
    }
    
    const adminBlacklist = ["root", "admin", "administrator", "superadmin", "super_admin", "admin_root", "root_admin", "system", "adminstrator"];
    if (adminBlacklist.includes(value.toLowerCase())) {
      setUsernameValid(false);
      setUsernameMessage("该用户名不允许注册");
      return;
    }
    
    const regex = /^[a-z0-9]+$/;
    if (!regex.test(value)) {
      setUsernameValid(false);
      setUsernameMessage("用户名只能包含小写字母和数字");
      return;
    }
    
    const exists = await checkExists("username", value);
    if (exists) {
      setUsernameValid(false);
      setUsernameMessage("系统已存在该用户名");
    } else {
      setUsernameValid(true);
      setUsernameMessage("用户名格式正确");
    }
  };

  const validateEmail = async (e) => {
    const value = e.target.value;
    setEmail(value);
    const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    
    if (value === "") {
      setEmailValid(true);
      setEmailMessage("");
      return;
    }
    
    if (!regex.test(value)) {
      setEmailValid(false);
      setEmailMessage("请输入有效的邮箱地址");
      return;
    }
    
    const exists = await checkExists("email", value);
    if (exists) {
      setEmailValid(false);
      setEmailMessage("系统已存在该邮箱");
    } else {
      setEmailValid(true);
      setEmailMessage("邮箱格式正确");
    }
  };

  const validatePhone = async (e) => {
    const value = e.target.value;
    setPhone(value);
    const regex = /^1[3-9]\d{9}$/;
    
    if (value === "") {
      setPhoneValid(true);
      setPhoneMessage("");
      return;
    }
    
    if (!regex.test(value)) {
      setPhoneValid(false);
      setPhoneMessage("请输入有效的手机号码");
      return;
    }
    
    const exists = await checkExists("phone", value);
    if (exists) {
      setPhoneValid(false);
      setPhoneMessage("系统已存在该手机号");
    } else {
      setPhoneValid(true);
      setPhoneMessage("手机号格式正确");
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    showMessage("", "success");

    if (!username || !phone || !email || !password || !confirmPassword) {
      showMessage("请填写所有字段", "error");
      return;
    }

    if (!usernameValid) {
      showMessage(usernameMessage, "error");
      return;
    }

    if (!phoneValid) {
      showMessage(phoneMessage, "error");
      return;
    }

    if (!emailValid) {
      showMessage(emailMessage, "error");
      return;
    }

    if (password !== confirmPassword) {
      showMessage("两次输入的密码不一致", "error");
      return;
    }

    if (!isPasswordValid(password)) {
      showMessage("密码不符合规范（需包含大小写字母、数字，且不少于6位）", "error");
      return;
    }

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
      const resp = await fetch("/api/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          username, 
          phone, 
          email, 
          password,
          captcha_token: captchaToken,
          captcha_code: captchaCode,
        }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "注册失败");
      }
      alert("注册成功，请登录");
      navigate("/login");
    } catch (err) {
      showMessage(err.message, "error");
      refreshCaptcha();
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refreshCaptcha();
  }, []);

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h1>欢迎注册MSE网关管理系统</h1>
        <h2>注册</h2>
        {message && <div className={`message ${messageType === 'error' ? 'message-error' : ''}`}>{message}</div>}
        <form onSubmit={handleSubmit}>
          <div>
            <label>用户名</label>
            <input
              type="text"
              value={username}
              onChange={validateUsername}
              required
              placeholder="请输入用户名（最少3个至多24个字符，小写字母和数字）"
              className={!usernameValid ? "invalid" : ""}
            />
            {usernameMessage && (
              <span className={`hint-text ${usernameValid ? "valid" : "invalid"}`}>
                {usernameMessage}
              </span>
            )}
          </div>
          <div>
            <label>手机号</label>
            <input
              type="tel"
              value={phone}
              onChange={validatePhone}
              required
              placeholder="请输入手机号"
              className={!phoneValid ? "invalid" : ""}
            />
            {phoneMessage && (
              <span className={`hint-text ${phoneValid ? "valid" : "invalid"}`}>
                {phoneMessage}
              </span>
            )}
          </div>
          <div>
            <label>邮箱</label>
            <input
              type="email"
              value={email}
              onChange={validateEmail}
              required
              placeholder="请输入邮箱"
              className={!emailValid ? "invalid" : ""}
            />
            {emailMessage && (
              <span className={`hint-text ${emailValid ? "valid" : "invalid"}`}>
                {emailMessage}
              </span>
            )}
          </div>
          <div>
            <label>密码</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="请输入密码（需包含大小写字母、数字，且不少于6位）"
            />
            <span className="hint-text">密码需包含大小写字母和数字，且不少于6位</span>
          </div>
          <div>
            <label>确认密码</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              placeholder="请再次输入密码"
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
            {loading ? "注册中..." : "注册"}
          </button>
        </form>
        <div className="auth-links">
          <span>已有账号？</span>
          <Link to="/login">立即登录</Link>
        </div>
      </div>
    </div>
  );
}
