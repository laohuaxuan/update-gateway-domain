import { useMemo, useRef, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";

const hasUppercase = /[A-Z]/;
const hasLowercase = /[a-z]/;
const hasDigit = /\d/;

const isPasswordValid = (password) =>
  password.length >= 6 && hasUppercase.test(password) && hasLowercase.test(password) && hasDigit.test(password);

export default function ResetPassword() {
  const [searchParams] = useSearchParams();
  const token = useMemo(() => searchParams.get("token") || "", [searchParams]);
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [message, setMessage] = useState("");
  const [messageType, setMessageType] = useState("success");
  const [loading, setLoading] = useState(false);
  const messageTimeoutRef = useRef(null);

  const showMessage = (text, type = "success") => {
    if (messageTimeoutRef.current) {
      clearTimeout(messageTimeoutRef.current);
    }
    setMessage(text);
    setMessageType(type);
    messageTimeoutRef.current = setTimeout(() => {
      setMessage("");
    }, 4000);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    showMessage("", "success");

    if (!token) {
      showMessage("重置链接无效，请重新发起找回密码。", "error");
      return;
    }
    if (!isPasswordValid(newPassword)) {
      showMessage("密码不符合规范（需包含大小写字母、数字，且不少于6位）。", "error");
      return;
    }
    if (newPassword !== confirmPassword) {
      showMessage("两次输入的密码不一致。", "error");
      return;
    }

    setLoading(true);
    try {
      const resp = await fetch("/api/change-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          token,
          new_password: newPassword,
        }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "密码重置失败");
      }
      setNewPassword("");
      setConfirmPassword("");
      showMessage("密码已重置成功，请返回登录。");
    } catch (err) {
      showMessage(err.message || "密码重置失败", "error");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h1>设置新密码</h1>
        <h2>找回密码</h2>
        {message && (
          <div className={`message ${messageType === "error" ? "message-error" : ""}`}>
            {message}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div>
            <label>新密码</label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
              placeholder="请输入新密码（需包含大小写字母、数字，且不少于6位）"
            />
            <span className="hint-text">密码需包含大小写字母和数字，且不少于6位</span>
          </div>
          <div>
            <label>确认新密码</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              placeholder="请再次输入新密码"
            />
          </div>
          <button type="submit" disabled={loading}>
            {loading ? "提交中..." : "确认修改"}
          </button>
        </form>

        <div className="auth-links">
          <Link to="/login">返回登录</Link>
        </div>
      </div>
    </div>
  );
}
