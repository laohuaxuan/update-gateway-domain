import { useEffect, useState, useRef } from "react";
import { useNavigate } from "react-router-dom";

const THEME_KEY = "update-gateway-domain-theme";

const navItems = [
  { id: "update", label: "更新数据", icon: "📦", roles: ["admin", "root"] },
  { id: "query", label: "查询数据", icon: "🔍", roles: ["watcher", "admin", "root"] },
  { id: "audit", label: "审计日志", icon: "📋", roles: ["admin", "root"] },
  { id: "users", label: "用户管理", icon: "👥", roles: ["root"] },
  { id: "datasource", label: "数据源", icon: "📊", roles: ["admin", "root"] },
];

const hasUppercase = /[A-Z]/;
const hasLowercase = /[a-z]/;
const hasDigit = /\d/;

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

export default function App() {
  const [accounts, setAccounts] = useState([]);
  const [selectedAccount, setSelectedAccount] = useState("");
  const [domains, setDomains] = useState([]);
  const [gateways, setGateways] = useState([]);
  const [selectedGateway, setSelectedGateway] = useState("");
  const [versions, setVersions] = useState(["tlsv1.0", "tlsv1.2", "tlsv1.3"]);
  const [domainName, setDomainName] = useState("");
  const [selectedCertIdentifier, setSelectedCertIdentifier] = useState("");
  const [gatewayCertificates, setGatewayCertificates] = useState([]);
  const [certDropdownOpen, setCertDropdownOpen] = useState(false);
  const [updateProtocol, setUpdateProtocol] = useState("DEFAULT");
  const [selectedVersion, setSelectedVersion] = useState("");
  const [batchVersion, setBatchVersion] = useState("");
  const [dryRun, setDryRun] = useState(false);
  const [mustHTTPS, setMustHTTPS] = useState("");
  const [http2, setHttp2] = useState("default");
  const [searchType, setSearchType] = useState("domain");
  const [searchKeyword, setSearchKeyword] = useState("");
  const [auditLogs, setAuditLogs] = useState([]);
  const [auditPage, setAuditPage] = useState(1);
  const [auditSize] = useState(10);
  const [auditTotal, setAuditTotal] = useState(0);
  const [auditGateway, setAuditGateway] = useState("");
  const [auditAction, setAuditAction] = useState("");
  const [auditResult, setAuditResult] = useState("");
  const [batchSummary, setBatchSummary] = useState(null);
  const [singleUpdateResult, setSingleUpdateResult] = useState(null);
  const [errorDialog, setErrorDialog] = useState(null);
  const [previewDialog, setPreviewDialog] = useState(null);
  const [passwordVisible, setPasswordVisible] = useState({
    old: false,
    next: false,
    confirm: false,
  });
  const [passwordModalMessage, setPasswordModalMessage] = useState("");
  const [passwordModalMessageType, setPasswordModalMessageType] = useState("success");
  const [passwordFieldErrors, setPasswordFieldErrors] = useState({
    old: "",
    next: "",
    confirm: "",
  });
  const [oldPasswordChecking, setOldPasswordChecking] = useState(false);
  const oldPasswordVerifySeqRef = useRef(0);
  const oldPasswordVerifyTimerRef = useRef(null);
  const [copyStatus, setCopyStatus] = useState("");
  const [message, setMessage] = useState("");
  const [messageType, setMessageType] = useState("success");
  const [loading, setLoading] = useState(false);
  const messageTimeoutRef = useRef(null);
  const [darkMode, setDarkMode] = useState(() => {
    try {
      const saved = localStorage.getItem(THEME_KEY);
      return saved === "dark";
    } catch {
      return false;
    }
  });
  const [domainPage, setDomainPage] = useState(1);
  const [domainSize, setDomainSize] = useState(20);
  const [domainTotal, setDomainTotal] = useState(0);
  const [users, setUsers] = useState([]);
  const [roleModal, setRoleModal] = useState(null);
  const [deleteModal, setDeleteModal] = useState(null);
  const [profileModal, setProfileModal] = useState(null);
  const [passwordModal, setPasswordModal] = useState(null);
  const [createUserModal, setCreateUserModal] = useState(false);
  const [newUser, setNewUser] = useState({ username: "", phone: "", email: "", role: "watcher" });
  const [resetPasswordModal, setResetPasswordModal] = useState(null);
  const [activeNav, setActiveNav] = useState("query");

  const [allAccounts, setAllAccounts] = useState([]);
  const [allGateways, setAllGateways] = useState([]);
  const [addAccountModal, setAddAccountModal] = useState(false);
  const [addGatewayModal, setAddGatewayModal] = useState(false);
  const [newAccount, setNewAccount] = useState({ name: "", id: "", region_id: "", access_key_id: "", access_key_secret: "", accept_language: "zh" });
  const [newGateway, setNewGateway] = useState({ account_id: "", name: "", id: "", batch_tls_version: "" });

  const [token, setToken] = useState(localStorage.getItem("token") || "");
  const [role, setRole] = useState(localStorage.getItem("role") || "");
  const [userName, setUserName] = useState(localStorage.getItem("name") || "");
  const [userEmail, setUserEmail] = useState(localStorage.getItem("email") || "");
  const [userPhone, setUserPhone] = useState(localStorage.getItem("phone") || "");

  const navigate = useNavigate();

  const isAdmin = role === "admin" || role === "root";
  const isRoot = role === "root";

  const authHeaders = {
    "Content-Type": "application/json",
    Authorization: token,
  };

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

  const openProfileModal = () => {
    setProfileModal({
      name: userName,
      phone: userPhone,
      email: userEmail,
    });
  };

  const openPasswordModal = () => {
    setPasswordVisible({ old: false, next: false, confirm: false });
    setPasswordModalMessage("");
    setPasswordModalMessageType("success");
    setPasswordFieldErrors({ old: "", next: "", confirm: "" });
    setOldPasswordChecking(false);
    oldPasswordVerifySeqRef.current += 1;
    if (oldPasswordVerifyTimerRef.current) {
      clearTimeout(oldPasswordVerifyTimerRef.current);
      oldPasswordVerifyTimerRef.current = null;
    }
    setPasswordModal({
      oldPassword: "",
      newPassword: "",
      confirmPassword: "",
    });
  };

  const closePasswordModal = () => {
    setPasswordModal(null);
    setPasswordModalMessage("");
    setPasswordFieldErrors({ old: "", next: "", confirm: "" });
    setOldPasswordChecking(false);
    oldPasswordVerifySeqRef.current += 1;
    if (oldPasswordVerifyTimerRef.current) {
      clearTimeout(oldPasswordVerifyTimerRef.current);
      oldPasswordVerifyTimerRef.current = null;
    }
  };

  const validateNewPassword = (password) => {
    if (!password) return "";
    if (!(password.length >= 6 && hasUppercase.test(password) && hasLowercase.test(password) && hasDigit.test(password))) {
      return "新密码不符合规范（需包含大小写字母、数字，且不少于6位）";
    }
    return "";
  };

  const validateConfirmPassword = (newPassword, confirmPassword) => {
    if (!confirmPassword) return "";
    if (newPassword !== confirmPassword) return "密码不一致";
    return "";
  };

  const verifyOldPassword = async (oldPassword) => {
    const resp = await fetch("/api/user-password/verify-old", {
      method: "POST",
      headers: authHeaders,
      body: JSON.stringify({ old_password: oldPassword }),
    });
    const data = await resp.json();
    if (!resp.ok) {
      throw new Error(data.error || "密码校验失败");
    }
    return !!data.valid;
  };

  const scheduleVerifyOldPassword = (oldPassword) => {
    if (oldPasswordVerifyTimerRef.current) {
      clearTimeout(oldPasswordVerifyTimerRef.current);
      oldPasswordVerifyTimerRef.current = null;
    }
    const seq = oldPasswordVerifySeqRef.current + 1;
    oldPasswordVerifySeqRef.current = seq;

    if (!oldPassword) {
      setOldPasswordChecking(false);
      setPasswordFieldErrors((prev) => ({ ...prev, old: "" }));
      return;
    }

    setOldPasswordChecking(true);
    oldPasswordVerifyTimerRef.current = setTimeout(async () => {
      try {
        const valid = await verifyOldPassword(oldPassword);
        if (oldPasswordVerifySeqRef.current !== seq) return;
        setPasswordFieldErrors((prev) => ({ ...prev, old: valid ? "" : "密码不正确" }));
      } catch {
        if (oldPasswordVerifySeqRef.current !== seq) return;
        setPasswordFieldErrors((prev) => ({ ...prev, old: "旧密码校验失败，请稍后重试" }));
      } finally {
        if (oldPasswordVerifySeqRef.current === seq) {
          setOldPasswordChecking(false);
        }
      }
    }, 400);
  };

  const checkAuth = async () => {
    if (!token) {
      navigate("/login");
      return;
    }
    try {
      const resp = await fetch("/api/user-info", { headers: { Authorization: token } });
      if (!resp.ok) {
        throw new Error("unauthorized");
      }
      const data = await resp.json();
      setRole(data.role || "");
      setUserName(data.name || "");
      setUserEmail(data.email || "");
      setUserPhone(data.phone || "");
      localStorage.setItem("role", data.role || "");
      localStorage.setItem("name", data.name || "");
      localStorage.setItem("email", data.email || "");
      localStorage.setItem("phone", data.phone || "");
    } catch {
      localStorage.removeItem("token");
      localStorage.removeItem("role");
      localStorage.removeItem("name");
      localStorage.removeItem("email");
      localStorage.removeItem("phone");
      setToken("");
      setRole("");
      navigate("/login");
    }
  };

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("role");
    localStorage.removeItem("name");
    localStorage.removeItem("email");
    localStorage.removeItem("phone");
    setToken("");
    setRole("");
    navigate("/login");
  };

  const loadDomains = async (nextPage = domainPage) => {
    if (!selectedAccount || selectedAccount === "all" || !selectedGateway) {
      setDomains([]);
      setDomainTotal(0);
      return;
    }
    setLoading(true);
    setMessage("");
    try {
      const params = new URLSearchParams();
      params.set("account_id", selectedAccount);
      params.set("gateway_id", selectedGateway);
      params.set("page", String(nextPage));
      params.set("size", String(domainSize));
      if (searchKeyword) {
        params.set("search_type", searchType);
        params.set("search_keyword", searchKeyword);
      }
      const resp = await fetch(`/api/domains?${params.toString()}`, { headers: { Authorization: token } });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "加载域名失败");
      }
      setDomains(data.domains || []);
      setDomainTotal(data.total || 0);
      setDomainPage(data.page || nextPage);
    } catch (err) {
      showMessage(err.message, "error");
    } finally {
      setLoading(false);
    }
  };

  const loadAccounts = async () => {
    try {
      const resp = await fetch("/api/accounts", { headers: { Authorization: token } });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "加载账号失败");
      }
      const list = data.accounts || [];
      setAccounts(list);
      if (list.length > 0) {
        setSelectedAccount((old) => old || list[0].id);
      }
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  const loadGateways = async () => {
    if (!selectedAccount || selectedAccount === "all") {
      setGateways([]);
      setSelectedGateway("");
      return;
    }
    try {
      const resp = await fetch(`/api/gateways?account_id=${encodeURIComponent(selectedAccount)}`, { headers: { Authorization: token } });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "加载网关失败");
      }
      const list = data.gateways || [];
      setGateways(list);
      if (list.length > 0) {
        setSelectedGateway((old) => {
          if (!old) return list[0].id;
          const exists = list.some((g) => g.id === old);
          return exists ? old : list[0].id;
        });
      } else {
        setSelectedGateway("");
      }
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  const loadGatewayCertificates = async () => {
    if (!selectedAccount || selectedAccount === "all" || !selectedGateway) {
      setGatewayCertificates([]);
      setSelectedCertIdentifier("");
      setCertDropdownOpen(false);
      return;
    }
    try {
      const params = new URLSearchParams();
      params.set("account_id", selectedAccount);
      params.set("gateway_id", selectedGateway);
      const resp = await fetch(`/api/gateway-certificates?${params.toString()}`, { headers: { Authorization: token } });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "加载证书失败");
      }
      setGatewayCertificates(data.certificates || []);
      setSelectedCertIdentifier("");
      setCertDropdownOpen(false);
    } catch (err) {
      setGatewayCertificates([]);
      setSelectedCertIdentifier("");
      setCertDropdownOpen(false);
      showMessage(err.message, "error");
    }
  };

  const loadAuditLogs = async (nextPage = auditPage) => {
    try {
      const params = new URLSearchParams();
      params.set("page", String(nextPage));
      params.set("size", String(auditSize));
      if (selectedAccount && selectedAccount !== "all") params.set("account_id", selectedAccount);
      if (auditGateway) params.set("gateway_id", auditGateway);
      if (auditAction) params.set("action", auditAction);
      if (auditResult) params.set("result", auditResult);
      const resp = await fetch(`/api/audit-logs?${params.toString()}`, { headers: { Authorization: token } });
      const data = await resp.json();
      if (resp.ok) {
        setAuditLogs(data.logs || []);
        setAuditTotal(data.total || 0);
        setAuditPage(data.page || nextPage);
      }
    } catch {
      // ignore
    }
  };

  const loadTLSVersions = async () => {
    try {
      const resp = await fetch("/api/tls-versions", { headers: { Authorization: token } });
      const data = await resp.json();
      if (resp.ok && Array.isArray(data.versions) && data.versions.length > 0) {
        setVersions(data.versions);
      }
    } catch {
      // ignore
    }
  };

  const loadUsers = async () => {
    if (!isRoot) return;
    try {
      const resp = await fetch("/api/users", { headers: authHeaders });
      const data = await resp.json();
      if (resp.ok) {
        setUsers(data.users || []);
      }
    } catch {
      // ignore
    }
  };

  const loadAllAccounts = async () => {
    if (!isAdmin) return;
    try {
      const resp = await fetch("/api/accounts/all", { headers: authHeaders });
      const data = await resp.json();
      if (resp.ok) {
        setAllAccounts(data.accounts || []);
      }
    } catch {
      // ignore
    }
  };

  const loadAllGateways = async () => {
    if (!isAdmin) return;
    try {
      const resp = await fetch("/api/gateways/all", { headers: authHeaders });
      const data = await resp.json();
      if (resp.ok) {
        setAllGateways(data.gateways || []);
      }
    } catch {
      // ignore
    }
  };

  const handleAddAccount = async () => {
    if (!newAccount.name || !newAccount.id || !newAccount.region_id || !newAccount.access_key_id || !newAccount.access_key_secret) {
      showMessage("请填写完整的账号信息", "error");
      return;
    }
    try {
      const resp = await fetch("/api/accounts", {
        method: "POST",
        headers: authHeaders,
        body: JSON.stringify(newAccount),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "添加账号失败");
      }
      showMessage("账号添加成功");
      setAddAccountModal(false);
      setNewAccount({ name: "", id: "", region_id: "", access_key_id: "", access_key_secret: "", accept_language: "zh-CN" });
      loadAllAccounts();
      loadAccounts();
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  const handleDeleteAccount = async (id) => {
    if (!confirm("确定要删除该账号吗？")) return;
    try {
      const resp = await fetch(`/api/accounts/${id}`, {
        method: "DELETE",
        headers: authHeaders,
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "删除账号失败");
      }
      showMessage("账号删除成功");
      loadAllAccounts();
      loadAccounts();
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  const handleAddGateway = async () => {
    if (!newGateway.account_id || !newGateway.name || !newGateway.id) {
      showMessage("请填写完整的网关信息", "error");
      return;
    }
    try {
      const resp = await fetch("/api/gateways", {
        method: "POST",
        headers: authHeaders,
        body: JSON.stringify(newGateway),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "添加网关失败");
      }
      showMessage("网关添加成功");
      setAddGatewayModal(false);
      setNewGateway({ account_id: "", name: "", id: "", batch_tls_version: "" });
      loadAllGateways();
      loadGateways();
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  const handleDeleteGateway = async (id) => {
    if (!confirm("确定要删除该网关吗？")) return;
    try {
      const resp = await fetch(`/api/gateways/${id}`, {
        method: "DELETE",
        headers: authHeaders,
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "删除网关失败");
      }
      showMessage("网关删除成功");
      loadAllGateways();
      loadGateways();
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  useEffect(() => {
    if (!token) return;
    loadAccounts();
    loadTLSVersions();
    loadUsers();
    loadAllAccounts();
    loadAllGateways();
    if (isAdmin) {
      loadAuditLogs(1);
    }
  }, [token]);

  useEffect(() => {
    document.body.classList.toggle("dark", darkMode);
    try {
      localStorage.setItem(THEME_KEY, darkMode ? "dark" : "light");
    } catch {
      // ignore
    }
  }, [darkMode]);

  useEffect(() => {
    setAuditGateway("");
    loadGateways();
  }, [selectedAccount]);

  useEffect(() => {
    setDomainPage(1);
    if (selectedGateway) {
      loadDomains(1);
    } else if (selectedAccount === "all") {
      setDomains([]);
      setDomainTotal(0);
    }
  }, [selectedAccount, selectedGateway, domainSize]);

  useEffect(() => {
    loadGatewayCertificates();
  }, [selectedAccount, selectedGateway, token]);

  useEffect(() => {
    if (isAdmin) {
      loadAuditLogs(1);
    }
  }, [selectedAccount, auditGateway, auditAction, auditResult]);

  useEffect(() => {
    if (activeNav === "datasource" && isAdmin) {
      loadAllAccounts();
      loadAllGateways();
    }
  }, [activeNav]);

  useEffect(() => {
    if (updateProtocol === "HTTP") {
      setMustHTTPS("false");
    }
  }, [updateProtocol]);

  const updateOneDomain = async () => {
    setLoading(true);
    showMessage("", "success");
    try {
      const resp = await fetch("/api/update-domain", {
        method: "POST",
        headers: authHeaders,
        body: JSON.stringify({
          account_id: selectedAccount,
          gateway_id: selectedGateway,
          domain: domainName,
          protocol: updateProtocol,
          tls_version: selectedVersion,
          cert_identifier: selectedCertIdentifier,
          dry_run: dryRun,
          must_https: mustHTTPS,
          http2: http2,
        }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "更新失败");
      }
      setSingleUpdateResult(data);
      showMessage(
        data.dry_run
          ? data.changed
            ? `Dry-run 预览完成：域名 ${data.domain || domainName} 检测到可变更项`
            : `Dry-run 预览完成：域名 ${data.domain || domainName} 无需变更`
          : data.changed
            ? `域名 ${data.domain || domainName} 更新成功`
            : `域名 ${data.domain || domainName} 无需变更`
      );
      await loadDomains();
      await loadAuditLogs(1);
    } catch (err) {
      setSingleUpdateResult(null);
      showMessage(err.message, "error");
    } finally {
      setLoading(false);
    }
  };

  const batchUpdate = async () => {
    setLoading(true);
    showMessage("", "success");
    try {
      const resp = await fetch("/api/batch-update", {
        method: "POST",
        headers: authHeaders,
        body: JSON.stringify({
          account_id: selectedAccount,
          gateway_id: selectedAccount === "all" ? "" : selectedGateway,
          protocol: updateProtocol,
          tls_version: batchVersion,
          dry_run: dryRun,
          must_https: mustHTTPS,
          http2: http2,
        }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "批量更新失败");
      }
      setBatchSummary(data);
      setErrorDialog(null);
      const targetText =
        data.account_id === "all"
          ? `全部账号/网关，共 ${data.results?.length || 0} 组`
          : `账号 ${data.account_id || "-"} 网关 ${data.gateway_id || "-"}`;
      showMessage(
        `${data.dry_run ? "Dry-run 预览完成" : "批量更新完成"}（${targetText}）：扫描 ${data.scanned_count || 0}，待更新 ${
          data.would_update_count || 0
        }，成功 ${data.updated_count || 0}，失败 ${data.failed_count || 0}`
      );
      if (selectedAccount !== "all") {
        await loadDomains();
      } else {
        setDomains([]);
      }
      await loadAuditLogs(1);
    } catch (err) {
      setBatchSummary(null);
      showMessage(err.message, "error");
    } finally {
      setLoading(false);
    }
  };

  const updateUserRole = async (userId, newRole) => {
    try {
      const resp = await fetch(`/api/users/${userId}/role`, {
        method: "PUT",
        headers: authHeaders,
        body: JSON.stringify({ role: newRole }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "更新角色失败");
      }
      showMessage("角色更新成功");
      setRoleModal(null);
      loadUsers();
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  const deleteUser = async (userId) => {
    try {
      const resp = await fetch(`/api/users/${userId}`, {
        method: "DELETE",
        headers: authHeaders,
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "删除用户失败");
      }
      showMessage("用户删除成功");
      setDeleteModal(null);
      loadUsers();
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  const changeOwnPassword = async () => {
    if (!passwordModal) return;
    setPasswordModalMessage("");
    setPasswordModalMessageType("success");
    const nextPasswordError = validateNewPassword(passwordModal.newPassword);
    const confirmPasswordError = validateConfirmPassword(passwordModal.newPassword, passwordModal.confirmPassword);
    setPasswordFieldErrors((prev) => ({
      ...prev,
      next: nextPasswordError,
      confirm: confirmPasswordError,
    }));
    if (!passwordModal.oldPassword || !passwordModal.newPassword || !passwordModal.confirmPassword) {
      setPasswordModalMessage("请完整填写旧密码、新密码和确认密码");
      setPasswordModalMessageType("error");
      return;
    }
    if (nextPasswordError) {
      setPasswordModalMessage(nextPasswordError);
      setPasswordModalMessageType("error");
      return;
    }
    if (passwordModal.oldPassword === passwordModal.newPassword) {
      setPasswordModalMessage("新密码不能与旧密码相同");
      setPasswordModalMessageType("error");
      return;
    }
    if (confirmPasswordError) {
      setPasswordModalMessage(confirmPasswordError);
      setPasswordModalMessageType("error");
      return;
    }
    if (oldPasswordChecking) {
      setPasswordModalMessage("正在校验旧密码，请稍候");
      setPasswordModalMessageType("error");
      return;
    }
    if (passwordFieldErrors.old) {
      setPasswordModalMessage(passwordFieldErrors.old);
      setPasswordModalMessageType("error");
      return;
    }
    try {
      const valid = await verifyOldPassword(passwordModal.oldPassword);
      if (!valid) {
        setPasswordFieldErrors((prev) => ({ ...prev, old: "密码不正确" }));
        setPasswordModalMessage("密码不正确");
        setPasswordModalMessageType("error");
        return;
      }
      const resp = await fetch("/api/user-password", {
        method: "PUT",
        headers: authHeaders,
        body: JSON.stringify({
          old_password: passwordModal.oldPassword,
          new_password: passwordModal.newPassword,
          confirm_password: passwordModal.confirmPassword,
        }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "修改密码失败");
      }
      setPasswordModalMessage("密码修改成功，请使用新密码重新登录");
      setPasswordModalMessageType("success");
      setTimeout(() => {
        closePasswordModal();
        handleLogout();
      }, 900);
    } catch (err) {
      setPasswordModalMessage(err.message || "修改密码失败");
      setPasswordModalMessageType("error");
    }
  };

  const handleCreateUser = async () => {
    if (!newUser.username || !newUser.phone || !newUser.email) {
      showMessage("请填写完整的用户信息", "error");
      return;
    }
    try {
      const resp = await fetch("/api/users", {
        method: "POST",
        headers: authHeaders,
        body: JSON.stringify(newUser),
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "创建用户失败");
      }
      alert(`用户创建成功！\n用户名：${newUser.username}\n密码：${data.password}\n（请务必保存此密码，只显示一次）`);
      showMessage("用户创建成功");
      setCreateUserModal(false);
      setNewUser({ username: "", phone: "", email: "", role: "watcher" });
      loadUsers();
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  const handleResetPassword = async (userId, userName) => {
    try {
      const resp = await fetch(`/api/users/${userId}/reset-password`, {
        method: "PUT",
        headers: authHeaders,
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || "重置密码失败");
      }
      alert(`密码重置成功！\n用户名：${userName}\n新密码：${data.password}\n（请务必保存此密码，只显示一次）`);
      showMessage("密码重置成功");
      setResetPasswordModal(null);
      loadUsers();
    } catch (err) {
      showMessage(err.message, "error");
    }
  };

  const batchRows = (() => {
    if (!batchSummary) return [];
    if (Array.isArray(batchSummary.results) && batchSummary.results.length > 0) {
      return batchSummary.results;
    }
    return [
      {
        account_id: batchSummary.account_id,
        gateway_id: batchSummary.gateway_id,
        requested_version: batchSummary.requested_version,
        scanned_count: batchSummary.scanned_count,
        would_update_count: batchSummary.would_update_count,
        updated_count: batchSummary.updated_count,
        failed_count: batchSummary.failed_count,
        errors: batchSummary.errors || [],
        changes: batchSummary.changes || [],
      },
    ];
  })();

  const formatMustHTTPS = (value) => (value === undefined || value === null ? "-" : value ? "开启" : "关闭");
  const formatHTTP2 = (value) => (value || "-");
  const formatTLSRange = (item) =>
    item.tls_min === item.tls_max ? item.tls_min : `${item.tls_min} / ${item.tls_max}`;
  const formatCertExpireAt = (value) => {
    if (!value) return "-";
    const matched = String(value).match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})/);
    return matched ? matched[1] : value;
  };
  const parseCertExpireDate = (value) => {
    if (!value) return null;
    const raw = String(value).trim();
    if (!raw) return null;
    const normalized = raw.replace(/([+-]\d{2})(\d{2})$/, "$1:$2");
    const parsed = new Date(normalized);
    if (Number.isNaN(parsed.getTime())) {
      return null;
    }
    return parsed;
  };
  const getCertExpiryProgress = (value) => {
    const expireDate = parseCertExpireDate(value);
    if (!expireDate) {
      return { level: "danger", percent: 0, daysLeft: null, isExpired: false, daysLeftText: "到期时间无效" };
    }
    const msPerDay = 24 * 60 * 60 * 1000;
    const daysLeft = Math.ceil((expireDate.getTime() - Date.now()) / msPerDay);
    let level = "safe";
    if (daysLeft <= 7) {
      level = "danger";
    } else if (daysLeft <= 30) {
      level = "warning";
    }
    const percent = Math.max(0, Math.min(100, Math.round((daysLeft / 30) * 100)));
    return {
      level,
      percent: daysLeft > 30 ? 100 : percent,
      daysLeft,
      isExpired: daysLeft < 0,
      daysLeftText: daysLeft < 0 ? `已过期 ${Math.abs(daysLeft)} 天` : `剩余 ${daysLeft} 天`,
    };
  };
  const getAuditChangeStatus = (item) => {
    if (item.message === "no-change") {
      return "无变更";
    }
    const hasBeforeAfter =
      item.cert_before !== undefined ||
      item.cert_after !== undefined ||
      item.protocol_before !== undefined ||
      item.protocol_after !== undefined ||
      item.tls_before !== undefined ||
      item.tls_after !== undefined ||
      item.must_https_before !== undefined ||
      item.must_https_after !== undefined ||
      item.http2_before !== undefined ||
      item.http2_after !== undefined;
    if (!hasBeforeAfter) {
      return "-";
    }
    const certBefore = item.cert_before ?? "";
    const certAfter = item.cert_after ?? "";
    const protocolBefore = item.protocol_before ?? item.protocol ?? "";
    const protocolAfter = item.protocol_after ?? item.protocol ?? "";
    const tlsBefore = item.tls_before ?? item.tls_version ?? "";
    const tlsAfter = item.tls_after ?? item.tls_version ?? "";
    const mustBefore = item.must_https_before;
    const mustAfter = item.must_https_after;
    const http2Before = item.http2_before ?? item.http2 ?? "";
    const http2After = item.http2_after ?? item.http2 ?? "";

    if (
      certBefore !== certAfter ||
      protocolBefore !== protocolAfter ||
      tlsBefore !== tlsAfter ||
      (mustBefore !== undefined && mustAfter !== undefined && mustBefore !== mustAfter) ||
      http2Before !== http2After
    ) {
      return "有变更";
    }
    return "无变更";
  };
  const formatProtocolPreview = (beforeProtocol, afterProtocol) => {
    const beforeValue = beforeProtocol || "-";
    const afterValue = afterProtocol || "-";
    const baseText = `${beforeValue} → ${afterValue}`;
    if (updateProtocol === "DEFAULT" && beforeValue === afterValue) {
      return `${baseText}（保持原值）`;
    }
    return baseText;
  };

  const copyErrors = async () => {
    if (!errorDialog || !Array.isArray(errorDialog.errors) || errorDialog.errors.length === 0) {
      return;
    }
    const text = [`[${errorDialog.title}]`, ...errorDialog.errors].join("\n");
    try {
      await navigator.clipboard.writeText(text);
      setCopyStatus("已复制");
      setTimeout(() => setCopyStatus(""), 1500);
    } catch {
      setCopyStatus("复制失败");
      setTimeout(() => setCopyStatus(""), 1500);
    }
  };

  const filteredNavItems = navItems.filter((item) => item.roles.includes(role));
  const httpsOptionsEnabled = updateProtocol !== "HTTP";
  const selectedCert = gatewayCertificates.find((cert) => cert.cert_identifier === selectedCertIdentifier) || null;
  const loadingHint = loading ? (
    <div className="loading-hint">
      <span className="loading-spinner" aria-hidden="true" />
      <span>数据加载中，请稍候...</span>
    </div>
  ) : null;

  if (!token) {
    return <div className="loading">加载中...</div>;
  }

  const renderUpdateData = () => (
    <div className="main-content">
      {message && <div className={`message ${messageType === 'error' ? 'message-error' : ''}`}>{message}</div>}
      {loadingHint}

      <div className="card">
        <h2>网关与模式</h2>
        <div className="row">
          <select value={selectedAccount} onChange={(e) => setSelectedAccount(e.target.value)}>
            <option value="all">全部账号（仅批量）</option>
            {accounts.map((acct) => (
              <option value={acct.id} key={acct.id}>
                {acct.name || acct.id} ({acct.id})
              </option>
            ))}
          </select>
          <select value={selectedGateway} onChange={(e) => setSelectedGateway(e.target.value)}>
            {gateways.map((gw) => (
              <option value={gw.id} key={gw.id}>
                {gw.name || gw.id} ({gw.id})
              </option>
            ))}
          </select>
          <label className="inline">
            <input type="checkbox" checked={dryRun} onChange={(e) => setDryRun(e.target.checked)} />
            Dry-run（仅预览，不落库）
          </label>
        </div>
      </div>

      <div className="card">
        <h2>单域名更新</h2>
        <div className="row">
          <input
            className="domain-name-input"
            type="text"
            placeholder="请输入域名，例如 api.example.com"
            value={domainName}
            onChange={(e) => setDomainName(e.target.value)}
          />
          <button
            disabled={loading || !domainName.trim() || !selectedAccount || selectedAccount === "all" || !selectedGateway}
            onClick={updateOneDomain}
          >
            检测并替换
          </button>
        </div>
        <div className="row top-gap">
          <div className="form-item cert-select-item">
            <span className="form-label">更新证书</span>
            <div className="cert-picker">
              <button
                type="button"
                className="cert-picker-trigger"
                onClick={() => setCertDropdownOpen((prev) => !prev)}
                disabled={!selectedAccount || selectedAccount === "all" || !selectedGateway}
              >
                {selectedCert
                  ? (selectedCert.cert_name || selectedCert.cert_identifier)
                  : "保持原证书（不变更）"}
                <span className="cert-picker-arrow">{certDropdownOpen ? "▴" : "▾"}</span>
              </button>
              {certDropdownOpen && (
                <div className="cert-picker-dropdown">
                  <button
                    type="button"
                    className={`cert-picker-option ${!selectedCertIdentifier ? "active" : ""}`}
                    onClick={() => {
                      setSelectedCertIdentifier("");
                      setCertDropdownOpen(false);
                    }}
                  >
                    <div className="cert-line-title">保持原证书（不变更）</div>
                  </button>
                  {gatewayCertificates.map((cert) => (
                    (() => {
                      const certExpiry = getCertExpiryProgress(cert.cert_expire_at);
                      return (
                        <button
                          type="button"
                          className={`cert-picker-option ${selectedCertIdentifier === cert.cert_identifier ? "active" : ""}`}
                          key={cert.cert_identifier}
                          onClick={() => {
                            setSelectedCertIdentifier(cert.cert_identifier);
                            setCertDropdownOpen(false);
                          }}
                        >
                          <div className="cert-line-title">
                            证书名称：{cert.cert_name || "-"}
                            {certExpiry.isExpired ? (
                              <span className="cert-status-expired">已过期</span>
                            ) : certExpiry.daysLeft !== null && certExpiry.daysLeft <= 7 ? (
                              <span className="cert-status-warning">即将过期（7天内）</span>
                            ) : null}
                          </div>
                      <div className="cert-line-sub">证书ID：{cert.cert_identifier || "-"}</div>
                      <div className="cert-line-sub">绑定域名：{cert.bound_domain || "-"}</div>
                      <div className="cert-line-sub">过期时间：{formatCertExpireAt(cert.cert_expire_at) || "-"}</div>
                        </button>
                      );
                    })()
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
        <div className="row top-gap">
          <label className="form-item">
            <span className="form-label">HTTP/HTTPS</span>
            <select value={updateProtocol} onChange={(e) => setUpdateProtocol(e.target.value)}>
              <option value="DEFAULT">默认（保持原协议）</option>
              <option value="HTTPS">HTTPS</option>
              <option value="HTTP">HTTP</option>
            </select>
          </label>
          <label className="form-item">
            <span className="form-label">TLS版本</span>
            <select
              value={selectedVersion}
              onChange={(e) => setSelectedVersion(e.target.value)}
              disabled={!httpsOptionsEnabled}
            >
              <option value="">默认（保持原配置）</option>
              {versions.map((v) => (
                <option value={v} key={v}>
                  {v}
                </option>
              ))}
            </select>
          </label>
          <label className="form-item">
            <span className="form-label">强制HTTPS</span>
            <select value={mustHTTPS} onChange={(e) => setMustHTTPS(e.target.value)} disabled={!httpsOptionsEnabled}>
              <option value="">默认（保持原配置）</option>
              <option value="true">开启</option>
              <option value="false">关闭</option>
            </select>
          </label>
          <label className="form-item">
            <span className="form-label">HTTP/2</span>
            <select value={http2} onChange={(e) => setHttp2(e.target.value)} disabled={!httpsOptionsEnabled}>
              <option value="default">默认（保持原配置）</option>
              <option value="open">开启</option>
              <option value="close">关闭</option>
            </select>
          </label>
        </div>
        {updateProtocol === "DEFAULT" && (
          <span className="hint-text">默认模式：保持域名原协议；HTTP 域名保持 HTTP，HTTPS 域名保持 HTTPS</span>
        )}
        {!httpsOptionsEnabled && (
          <span className="hint-text">当前为 HTTP，TLS/强制HTTPS/HTTP/2 配置不可调整，且强制HTTPS固定为关闭</span>
        )}
        {singleUpdateResult?.preview && (
          <div className="top-gap">
            <h3>单域名变更预览</h3>
            <table>
              <thead>
                <tr>
                  <th>域名</th>
                  <th>协议</th>
                  <th>TLS 版本</th>
                  <th>强制HTTPS</th>
                  <th>HTTP/2</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>{singleUpdateResult.preview.domain}</td>
                  <td>
                    {formatProtocolPreview(singleUpdateResult.preview.before.protocol, singleUpdateResult.preview.after.protocol)}
                  </td>
                  <td>
                    {formatTLSRange(singleUpdateResult.preview.before)} →{" "}
                    {formatTLSRange(singleUpdateResult.preview.after)}
                  </td>
                  <td>
                    {formatMustHTTPS(singleUpdateResult.preview.before.must_https)} →{" "}
                    {formatMustHTTPS(singleUpdateResult.preview.after.must_https)}
                  </td>
                  <td>
                    {formatHTTP2(singleUpdateResult.preview.before.http2)} →{" "}
                    {formatHTTP2(singleUpdateResult.preview.after.http2)}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="card">
        <h2>批量更新</h2>
        <p>
          默认保持各域名当前 TLS/强制HTTPS/HTTP/2 配置，也可临时指定版本。
          当账号选择“全部账号”时，会自动遍历所有账号下所有网关执行。
        </p>
        <div className="row top-gap">
          <label className="form-item">
            <span className="form-label">HTTP/HTTPS</span>
            <select value={updateProtocol} onChange={(e) => setUpdateProtocol(e.target.value)}>
              <option value="DEFAULT">默认（保持原协议）</option>
              <option value="HTTPS">HTTPS</option>
              <option value="HTTP">HTTP</option>
            </select>
          </label>
          <label className="form-item">
            <span className="form-label">TLS版本</span>
            <select value={batchVersion} onChange={(e) => setBatchVersion(e.target.value)} disabled={!httpsOptionsEnabled}>
              <option value="">默认（保持原配置）</option>
              {versions.map((v) => (
                <option value={v} key={v}>
                  {v}
                </option>
              ))}
            </select>
          </label>
          <label className="form-item">
            <span className="form-label">强制HTTPS</span>
            <select value={mustHTTPS} onChange={(e) => setMustHTTPS(e.target.value)} disabled={!httpsOptionsEnabled}>
              <option value="">默认（保持原配置）</option>
              <option value="true">开启</option>
              <option value="false">关闭</option>
            </select>
          </label>
          <label className="form-item">
            <span className="form-label">HTTP/2</span>
            <select value={http2} onChange={(e) => setHttp2(e.target.value)} disabled={!httpsOptionsEnabled}>
              <option value="default">默认（保持原配置）</option>
              <option value="open">开启</option>
              <option value="close">关闭</option>
            </select>
          </label>
        </div>
        {updateProtocol === "DEFAULT" && (
          <span className="hint-text">默认模式：保持域名原协议；HTTP 域名保持 HTTP，HTTPS 域名保持 HTTPS</span>
        )}
        {!httpsOptionsEnabled && (
          <span className="hint-text">当前为 HTTP，TLS/强制HTTPS/HTTP/2 配置不可调整，且强制HTTPS固定为关闭</span>
        )}
        <button
          disabled={loading || !selectedAccount || (selectedAccount !== "all" && !selectedGateway)}
          onClick={batchUpdate}
          className="top-gap"
        >
          执行批量更新
        </button>
      </div>

      {batchSummary && (
        <div className="card">
          <div className="table-head">
            <h2>批量执行明细</h2>
            <span>
              总扫描 {batchSummary.scanned_count || 0} / 待更新 {batchSummary.would_update_count || 0} / 成功{" "}
              {batchSummary.updated_count || 0} / 失败 {batchSummary.failed_count || 0}
            </span>
          </div>
          <table>
            <thead>
              <tr>
                <th>账号</th>
                <th>网关</th>
                <th>TLS</th>
                <th>扫描数</th>
                <th>待更新</th>
                <th>成功</th>
                <th>失败</th>
                <th>变更预览</th>
                <th>错误</th>
              </tr>
            </thead>
            <tbody>
              {batchRows.map((row, idx) => (
                <tr key={`${row.account_id || "-"}-${row.gateway_id || "-"}-${idx}`}>
                  <td>{row.account_id || "-"}</td>
                  <td>{row.gateway_id || "-"}</td>
                  <td>{row.requested_version || "-"}</td>
                  <td>{row.scanned_count || 0}</td>
                  <td>{row.would_update_count || 0}</td>
                  <td>{row.updated_count || 0}</td>
                  <td>{row.failed_count || 0}</td>
                  <td>
                    {(row.changes || []).length > 0 ? (
                      <button
                        type="button"
                        className="link-btn"
                        onClick={() =>
                          setPreviewDialog({
                            title: `账号 ${row.account_id || "-"} / 网关 ${row.gateway_id || "-"}`,
                            changes: row.changes || [],
                          })
                        }
                      >
                        查看（{(row.changes || []).length}）
                      </button>
                    ) : (
                      "-"
                    )}
                  </td>
                  <td className="error-text">
                    {(row.errors || []).length > 0 ? (
                      <button
                        type="button"
                        className="link-btn"
                        onClick={() =>
                          setErrorDialog({
                            title: `账号 ${row.account_id || "-"} / 网关 ${row.gateway_id || "-"}`,
                            errors: row.errors || [],
                          })
                        }
                      >
                        查看详情（{(row.errors || []).length}）
                      </button>
                    ) : (
                      "-"
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );

  const renderQueryData = () => (
    <div className="main-content">
      {message && <div className={`message ${messageType === 'error' ? 'message-error' : ''}`}>{message}</div>}
      {loadingHint}

      <div className="card">
        <h2>网关选择</h2>
        <div className="row">
          <select value={selectedAccount} onChange={(e) => setSelectedAccount(e.target.value)}>
            {accounts.map((acct) => (
              <option value={acct.id} key={acct.id}>
                {acct.name || acct.id} ({acct.id})
              </option>
            ))}
          </select>
          <select value={selectedGateway} onChange={(e) => setSelectedGateway(e.target.value)}>
            {gateways.map((gw) => (
              <option value={gw.id} key={gw.id}>
                {gw.name || gw.id} ({gw.id})
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="card">
        <div className="table-head">
          <h2>网关域名列表</h2>
        </div>
        <div className="search-box">
          <div className="search-filter">
            <select value={searchType} onChange={(e) => setSearchType(e.target.value)}>
              <option value="domain">域名</option>
              <option value="protocol">协议</option>
              <option value="tls">TLS版本</option>
              <option value="cert">证书名称</option>
            </select>
            <button className="filter-clear" onClick={() => {
              setSearchType("domain");
              setSearchKeyword("");
              loadDomains(1);
            }}>×</button>
          </div>
          <input
            type="text"
            placeholder={`请输入${searchType === 'domain' ? '域名' : searchType === 'protocol' ? '协议' : searchType === 'tls' ? 'TLS版本' : '证书名称'}关键词`}
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && loadDomains(1)}
          />
          <button className="search-btn" disabled={loading} onClick={() => loadDomains(1)}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
          </button>
        </div>
        <div className="row top-gap">
          <select value={domainSize} onChange={(e) => setDomainSize(Number(e.target.value))}>
            <option value={10}>10条/页</option>
            <option value={20}>20条/页</option>
            <option value={50}>50条/页</option>
            <option value={100}>100条/页</option>
          </select>
          <button disabled={loading} onClick={() => loadDomains(1)}>
            刷新
          </button>
        </div>
        <table>
          <thead>
            <tr>
              <th>编号</th>
              <th>ID</th>
              <th>域名</th>
              <th>协议</th>
              <th>TLS Min</th>
              <th>TLS Max</th>
              <th>强制HTTPS</th>
              <th>开启HTTP/2</th>
              <th>证书名称</th>
              <th>证书过期时间</th>
            </tr>
          </thead>
          <tbody>
            {domains.map((d, idx) => {
              const certExpiry = getCertExpiryProgress(d.cert_expire_at);
              return (
              <tr key={d.id} className={certExpiry.isExpired ? "expired-row" : ""}>
                <td>{(domainPage - 1) * domainSize + idx + 1}</td>
                <td>{d.id}</td>
                <td>{d.name}</td>
                <td>{d.protocol}</td>
                <td>{d.tls_min}</td>
                <td>{d.tls_max}</td>
                <td>{d.must_https ? "✅" : "❌"}</td>
                <td>{d.http2 === "on" ? "✅" : "❌"}</td>
                <td>{d.cert_name || d.cert_identifier || "-"}</td>
                <td className="cert-expire-cell">
                  {d.cert_expire_at ? (
                    <>
                      <div>{formatCertExpireAt(d.cert_expire_at)}</div>
                      <div className="cert-expire-progress">
                        <div
                          className={`cert-expire-progress-fill cert-expire-progress-${certExpiry.level}`}
                          style={{ width: `${certExpiry.percent}%` }}
                        />
                      </div>
                      <span className={`cert-expire-days cert-expire-days-${certExpiry.level}`}>
                        {certExpiry.daysLeftText}
                      </span>
                    </>
                  ) : (
                    "-"
                  )}
                </td>
              </tr>
              );
            })}
          </tbody>
        </table>
        <div className="row top-gap">
          <button
            disabled={loading || domainPage <= 1}
            onClick={() => loadDomains(domainPage - 1)}
          >
            上一页
          </button>
          <button
            disabled={loading || domainPage * domainSize >= domainTotal}
            onClick={() => loadDomains(domainPage + 1)}
          >
            下一页
          </button>
          <span>第 {domainPage} 页 / 共 {Math.max(1, Math.ceil(domainTotal / domainSize))} 页</span>
          <span>总记录 {domainTotal}</span>
        </div>
      </div>
    </div>
  );

  const renderAuditLogs = () => (
    <div className="main-content">
      {message && <div className={`message ${messageType === 'error' ? 'message-error' : ''}`}>{message}</div>}
      {loadingHint}

      <div className="card">
        <div className="table-head">
          <h2>最近审计日志</h2>
          <button disabled={loading} onClick={() => loadAuditLogs(auditPage)}>
            刷新
          </button>
        </div>
        <div className="row">
          <select value={auditGateway} onChange={(e) => setAuditGateway(e.target.value)}>
            <option value="">全部网关</option>
            {gateways.map((gw) => (
              <option value={gw.id} key={`audit-${gw.id}`}>
                {gw.name || gw.id}
              </option>
            ))}
          </select>
          <select value={auditAction} onChange={(e) => setAuditAction(e.target.value)}>
            <option value="">全部动作</option>
            <option value="single_update">single_update</option>
            <option value="batch_update">batch_update</option>
          </select>
          <select value={auditResult} onChange={(e) => setAuditResult(e.target.value)}>
            <option value="">全部结果</option>
            <option value="success">success</option>
            <option value="failed">failed</option>
          </select>
        </div>
        <table>
          <thead>
            <tr>
              <th>时间</th>
              <th>账号</th>
              <th>动作</th>
              <th>网关</th>
              <th>域名</th>
              <th>是否变更</th>
              <th>证书（前→后）</th>
              <th>HTTP/HTTPS（前→后）</th>
              <th>TLS（前→后）</th>
              <th>强制HTTPS（前→后）</th>
              <th>HTTP/2（前→后）</th>
              <th>Dry-run</th>
              <th>结果</th>
            </tr>
          </thead>
          <tbody>
            {auditLogs.map((item, idx) => (
              <tr key={`${item.time}-${idx}`}>
                <td>{item.time}</td>
                <td>{item.account_id || "-"}</td>
                <td>{item.action}</td>
                <td>{item.gateway_id}</td>
                <td>{item.domain || "-"}</td>
                <td>{getAuditChangeStatus(item)}</td>
                <td>{`${item.cert_before || "-"} → ${item.cert_after || "-"}`}</td>
                <td>{`${item.protocol_before || item.protocol || "-"} → ${item.protocol_after || item.protocol || "-"}`}</td>
                <td>{`${item.tls_before || item.tls_version || "-"} → ${item.tls_after || item.tls_version || "-"}`}</td>
                <td>{`${formatMustHTTPS(item.must_https_before)} → ${formatMustHTTPS(item.must_https_after)}`}</td>
                <td>{`${formatHTTP2(item.http2_before)} → ${formatHTTP2(item.http2_after)}`}</td>
                <td>{String(item.dry_run)}</td>
                <td>{item.result}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <div className="row top-gap">
          <button
            disabled={loading || auditPage <= 1}
            onClick={() => loadAuditLogs(auditPage - 1)}
          >
            上一页
          </button>
          <button
            disabled={loading || auditPage * auditSize >= auditTotal}
            onClick={() => loadAuditLogs(auditPage + 1)}
          >
            下一页
          </button>
          <span>第 {auditPage} 页 / 共 {Math.max(1, Math.ceil(auditTotal / auditSize))} 页</span>
          <span>总记录 {auditTotal}</span>
        </div>
      </div>
    </div>
  );

  const renderUserManagement = () => (
    <div className="main-content">
      {message && <div className={`message ${messageType === 'error' ? 'message-error' : ''}`}>{message}</div>}
      {loadingHint}

      <div className="card">
        <div className="table-head">
          <h2>用户管理</h2>
          <div className="row">
            <button onClick={() => { setCreateUserModal(true); }}>创建用户</button>
            <button onClick={loadUsers}>刷新</button>
          </div>
        </div>
        <table>
          <thead>
            <tr>
              <th>姓名</th>
              <th>手机号</th>
              <th>邮箱</th>
              <th>角色</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id}>
                <td>{u.name}</td>
                <td>{u.phone}</td>
                <td>{u.email}</td>
                <td>{u.role === "root" ? "超级管理员" : u.role === "admin" ? "管理员" : "观察者"}</td>
                <td>
                  {u.role !== "root" && (
                    <>
                      <button
                      type="button"
                      className="link-btn"
                      onClick={() => setRoleModal({ id: u.id, name: u.name, role: u.role })}
                    >
                      变更角色
                    </button>
                    <span className="btn-spacing">|</span>
                    <button
                      type="button"
                      className="link-btn"
                      onClick={() => setResetPasswordModal({ id: u.id, name: u.name })}
                    >
                      重置密码
                    </button>
                    <span className="btn-spacing">|</span>
                    <button
                      type="button"
                      className="link-btn error-text"
                      onClick={() => setDeleteModal({ id: u.id, name: u.name })}
                    >
                      删除
                    </button>
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );

  const renderDataSource = () => (
    <div className="main-content">
      {message && <div className={`message ${messageType === 'error' ? 'message-error' : ''}`}>{message}</div>}
      {loadingHint}

      <div className="card">
        <div className="table-head">
          <h2>阿里云账号列表</h2>
          <button onClick={() => { setAddAccountModal(true); setAddGatewayModal(false); }}>添加账号</button>
        </div>
        <table>
          <thead>
            <tr>
              <th>账号名称</th>
              <th>账号ID</th>
              <th>地域</th>
              <th>AccessKey ID</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {allAccounts.map((acct) => (
              <tr key={acct.id}>
                <td>{acct.name}</td>
                <td>{acct.id}</td>
                <td>{acct.region_id}</td>
                <td>{acct.access_key_id}</td>
                <td>
                  <button
                    type="button"
                    className="link-btn"
                    onClick={() => {
                      setNewGateway({ ...newGateway, account_id: acct.id });
                      setAddGatewayModal(true);
                      setAddAccountModal(false);
                    }}
                  >
                    添加网关
                  </button>
                  <span className="btn-spacing">|</span>
                  <button
                    type="button"
                    className="link-btn error-text"
                    onClick={() => handleDeleteAccount(acct.id)}
                  >
                    删除
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="card">
        <div className="table-head">
          <h2>网关信息列表</h2>
        </div>
        <table>
          <thead>
            <tr>
              <th>网关名称</th>
              <th>网关ID</th>
              <th>所属账号</th>
              <th>批量TLS版本</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {allGateways.map((gw) => {
              const account = allAccounts.find((a) => a.id === gw.account_id);
              return (
                <tr key={gw.id}>
                  <td>{gw.name}</td>
                  <td>{gw.id}</td>
                  <td>{account?.name || gw.account_id}</td>
                  <td>{gw.batch_tls || "-"}</td>
                  <td>
                    <button
                      type="button"
                      className="link-btn error-text"
                      onClick={() => handleDeleteGateway(gw.id)}
                    >
                      删除
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {addAccountModal && (
        <div className="modal-overlay" onClick={() => setAddAccountModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>添加阿里云账号</h3>
            <div className="form-group">
              <label>账号名称：</label>
              <input type="text" value={newAccount.name} onChange={(e) => setNewAccount({ ...newAccount, name: e.target.value })} />
            </div>
            <div className="form-group">
              <label>账号ID：</label>
              <input type="text" value={newAccount.id} onChange={(e) => setNewAccount({ ...newAccount, id: e.target.value })} />
            </div>
            <div className="form-group">
              <label>地域ID：</label>
              <input type="text" value={newAccount.region_id} onChange={(e) => setNewAccount({ ...newAccount, region_id: e.target.value })} placeholder="如 cn-beijing" />
            </div>
            <div className="form-group">
              <label>AccessKey ID：</label>
              <input type="text" value={newAccount.access_key_id} onChange={(e) => setNewAccount({ ...newAccount, access_key_id: e.target.value })} />
            </div>
            <div className="form-group">
              <label>AccessKey Secret：</label>
              <input type="password" value={newAccount.access_key_secret} onChange={(e) => setNewAccount({ ...newAccount, access_key_secret: e.target.value })} />
            </div>
            <div className="form-group">
              <label>语言：</label>
              <select value={newAccount.accept_language} onChange={(e) => setNewAccount({ ...newAccount, accept_language: e.target.value })}>
                <option value="zh-CN">中文</option>
                <option value="en-US">英文</option>
              </select>
            </div>
            <div className="row">
              <button onClick={() => setAddAccountModal(false)}>取消</button>
              <button onClick={handleAddAccount}>确认添加</button>
            </div>
          </div>
        </div>
      )}

      {addGatewayModal && (
        <div className="modal-overlay" onClick={() => setAddGatewayModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>添加网关信息</h3>
            <div className="form-group">
              <label>所属账号：</label>
              <select value={newGateway.account_id} onChange={(e) => setNewGateway({ ...newGateway, account_id: e.target.value })}>
                <option value="">请选择账号</option>
                {allAccounts.map((acct) => (
                  <option value={acct.id} key={acct.id}>{acct.name} ({acct.id})</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>网关名称：</label>
              <input type="text" value={newGateway.name} onChange={(e) => setNewGateway({ ...newGateway, name: e.target.value })} />
            </div>
            <div className="form-group">
              <label>网关ID：</label>
              <input type="text" value={newGateway.id} onChange={(e) => setNewGateway({ ...newGateway, id: e.target.value })} />
            </div>
            <div className="form-group">
              <label>批量TLS版本（可选）：</label>
              <select value={newGateway.batch_tls_version} onChange={(e) => setNewGateway({ ...newGateway, batch_tls_version: e.target.value })}>
                <option value="">默认版本</option>
                {versions.map((v) => (
                  <option value={v} key={v}>{v}</option>
                ))}
              </select>
            </div>
            <div className="row">
              <button onClick={() => setAddGatewayModal(false)}>取消</button>
              <button onClick={handleAddGateway}>确认添加</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );

  const renderContent = () => {
    switch (activeNav) {
      case "update":
        return renderUpdateData();
      case "query":
        return renderQueryData();
      case "audit":
        return renderAuditLogs();
      case "users":
        return renderUserManagement();
      case "datasource":
        return renderDataSource();
      default:
        return renderQueryData();
    }
  };

  return (
    <div className={`app-container ${darkMode ? "dark" : ""}`}>
      <aside className="sidebar">
        <div className="sidebar-header">
          <h1>MSE 网关管理</h1>
        </div>
        <nav className="sidebar-nav">
          {filteredNavItems.map((item) => (
            <button
              key={item.id}
              className={`nav-item ${activeNav === item.id ? "active" : ""}`}
              onClick={() => setActiveNav(item.id)}
            >
              <span className="nav-icon">{item.icon}</span>
              <span className="nav-label">{item.label}</span>
            </button>
          ))}
        </nav>
      </aside>

      <div className="main-area">
        <header className="topbar">
          <div className="topbar-left">
            <h2>{filteredNavItems.find((item) => item.id === activeNav)?.label || "查询数据"}</h2>
          </div>
          <div className="topbar-right">
            <div className="theme-toggle">
              <span>{darkMode ? "🌙" : "☀️"}</span>
              <button type="button" onClick={() => setDarkMode(!darkMode)} className="theme-btn">
                {darkMode ? "明亮" : "暗黑"}
              </button>
            </div>
            <div className="user-menu">
              <button
                className="user-btn"
                onClick={openProfileModal}
              >
                <span className="avatar">
                  {userName ? userName.charAt(0).toUpperCase() : "U"}
                </span>
                <span className="user-details">
                  <span className="user-name">{userName}</span>
                  <span className="user-role">
                    {role === "root" ? "超级管理员" : role === "admin" ? "管理员" : "观察者"}
                  </span>
                </span>
              </button>
              <div className="user-dropdown">
                <button
                  className="dropdown-item"
                  onClick={openProfileModal}
                >
                  👤 个人设置
                </button>
                <button className="dropdown-item logout-btn" onClick={handleLogout}>
                  🚪 退出登录
                </button>
              </div>
            </div>
          </div>
        </header>

        {renderContent()}
      </div>

      {roleModal && (
        <div className="modal-mask" onClick={() => setRoleModal(null)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <h2>变更角色 - {roleModal.name}</h2>
            <select
              value={roleModal.role}
              onChange={(e) => updateUserRole(roleModal.id, e.target.value)}
            >
              <option value="watcher">观察者</option>
              <option value="admin">管理员</option>
            </select>
            <button onClick={() => setRoleModal(null)}>取消</button>
          </div>
        </div>
      )}

      {deleteModal && (
        <div className="modal-mask" onClick={() => setDeleteModal(null)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <h2>确认删除用户</h2>
            <p>确定要删除用户 {deleteModal.name} 吗？此操作不可撤销。</p>
            <div className="row">
              <button onClick={() => deleteUser(deleteModal.id)} className="error-btn">
                确认删除
              </button>
              <button onClick={() => setDeleteModal(null)}>取消</button>
            </div>
          </div>
        </div>
      )}

      {errorDialog && (
        <div className="modal-mask" onClick={() => setErrorDialog(null)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="table-head">
              <h2>错误详情</h2>
              <div className="row">
                <button type="button" className="link-btn action-btn" onClick={copyErrors}>
                  复制全部错误
                </button>
                <button type="button" onClick={() => setErrorDialog(null)}>
                  关闭
                </button>
              </div>
            </div>
            <p>{errorDialog.title}</p>
            {copyStatus && <p className="copy-status">{copyStatus}</p>}
            <ul className="error-list">
              {(errorDialog.errors || []).map((err, idx) => (
                <li key={`${idx}-${err}`}>{err}</li>
              ))}
            </ul>
          </div>
        </div>
      )}

      {previewDialog && (
        <div className="modal-mask" onClick={() => setPreviewDialog(null)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="table-head">
              <h2>Dry-run 变更预览</h2>
              <button type="button" onClick={() => setPreviewDialog(null)}>
                关闭
              </button>
            </div>
            <p>{previewDialog.title}</p>
            <table>
              <thead>
                <tr>
                  <th>域名</th>
                  <th>协议</th>
                  <th>TLS 版本</th>
                  <th>强制HTTPS</th>
                  <th>HTTP/2</th>
                </tr>
              </thead>
              <tbody>
                {(previewDialog.changes || []).map((change, idx) => (
                  <tr key={`${change.domain}-${idx}`}>
                    <td>{change.domain}</td>
                    <td>
                      {formatProtocolPreview(change.before.protocol, change.after.protocol)}
                    </td>
                    <td>
                      {formatTLSRange(change.before)} → {formatTLSRange(change.after)}
                    </td>
                    <td>
                      {formatMustHTTPS(change.before.must_https)} → {formatMustHTTPS(change.after.must_https)}
                    </td>
                    <td>
                      {formatHTTP2(change.before.http2)} → {formatHTTP2(change.after.http2)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {profileModal && (
        <div className="modal-mask" onClick={() => setProfileModal(null)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <h2>个人设置</h2>
            <div className="form-group">
              <label>姓名</label>
              <input
                type="text"
                value={profileModal.name}
                readOnly
                className="readonly-input"
              />
            </div>
            <div className="form-group">
              <label>手机号</label>
              <input
                type="text"
                value={profileModal.phone}
                readOnly
                className="readonly-input"
              />
            </div>
            <div className="form-group">
              <label>邮箱</label>
              <input
                type="email"
                value={profileModal.email}
                readOnly
                className="readonly-input"
              />
            </div>
            <div className="row">
              <button onClick={openPasswordModal}>修改账号密码</button>
              <button onClick={() => setProfileModal(null)}>取消</button>
            </div>
          </div>
        </div>
      )}

      {passwordModal && (
        <div className="modal-mask" onClick={closePasswordModal}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <h2>修改账号密码</h2>
            {passwordModalMessage && (
              <div className={`message ${passwordModalMessageType === "error" ? "message-error" : ""}`}>
                {passwordModalMessage}
              </div>
            )}
            <div className="form-group">
              <label>旧密码</label>
              <div className="password-input-row">
                <input
                  type={passwordVisible.old ? "text" : "password"}
                  value={passwordModal.oldPassword}
                  onChange={(e) => {
                    const next = e.target.value;
                    setPasswordModal({ ...passwordModal, oldPassword: next });
                    scheduleVerifyOldPassword(next);
                  }}
                  placeholder="请输入当前密码"
                />
                <button
                  type="button"
                  className="password-toggle-btn"
                  onClick={() => setPasswordVisible((prev) => ({ ...prev, old: !prev.old }))}
                  title={passwordVisible.old ? "隐藏密码" : "显示密码"}
                  aria-label={passwordVisible.old ? "隐藏密码" : "显示密码"}
                >
                  <EyeToggleIcon visible={passwordVisible.old} />
                </button>
              </div>
              {oldPasswordChecking ? (
                <span className="hint-text">正在校验旧密码...</span>
              ) : (
                passwordFieldErrors.old && <span className="hint-text invalid">{passwordFieldErrors.old}</span>
              )}
            </div>
            <div className="form-group">
              <label>新密码</label>
              <div className="password-input-row">
                <input
                  type={passwordVisible.next ? "text" : "password"}
                  value={passwordModal.newPassword}
                  onChange={(e) => {
                    const next = e.target.value;
                    const nextErr = validateNewPassword(next);
                    const confirmErr = validateConfirmPassword(next, passwordModal.confirmPassword);
                    setPasswordModal({ ...passwordModal, newPassword: next });
                    setPasswordFieldErrors((prev) => ({ ...prev, next: nextErr, confirm: confirmErr }));
                  }}
                  placeholder="请输入新密码（需包含大小写字母、数字，且不少于6位）"
                />
                <button
                  type="button"
                  className="password-toggle-btn"
                  onClick={() => setPasswordVisible((prev) => ({ ...prev, next: !prev.next }))}
                  title={passwordVisible.next ? "隐藏密码" : "显示密码"}
                  aria-label={passwordVisible.next ? "隐藏密码" : "显示密码"}
                >
                  <EyeToggleIcon visible={passwordVisible.next} />
                </button>
              </div>
              {passwordFieldErrors.next && <span className="hint-text invalid">{passwordFieldErrors.next}</span>}
              <span className="hint-text">密码需包含大小写字母和数字，且不少于6位</span>
            </div>
            <div className="form-group">
              <label>确认新密码</label>
              <div className="password-input-row">
                <input
                  type={passwordVisible.confirm ? "text" : "password"}
                  value={passwordModal.confirmPassword}
                  onChange={(e) => {
                    const next = e.target.value;
                    const confirmErr = validateConfirmPassword(passwordModal.newPassword, next);
                    setPasswordModal({ ...passwordModal, confirmPassword: next });
                    setPasswordFieldErrors((prev) => ({ ...prev, confirm: confirmErr }));
                  }}
                  placeholder="请再次输入新密码"
                />
                <button
                  type="button"
                  className="password-toggle-btn"
                  onClick={() => setPasswordVisible((prev) => ({ ...prev, confirm: !prev.confirm }))}
                  title={passwordVisible.confirm ? "隐藏密码" : "显示密码"}
                  aria-label={passwordVisible.confirm ? "隐藏密码" : "显示密码"}
                >
                  <EyeToggleIcon visible={passwordVisible.confirm} />
                </button>
              </div>
              {passwordFieldErrors.confirm && <span className="hint-text invalid">{passwordFieldErrors.confirm}</span>}
            </div>
            <div className="row">
              <button onClick={changeOwnPassword}>确认修改</button>
              <button onClick={closePasswordModal}>取消</button>
            </div>
          </div>
        </div>
      )}

      {createUserModal && (
        <div className="modal-mask" onClick={() => setCreateUserModal(false)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <h2>创建新用户</h2>
            <div className="form-group">
              <label>用户名（最少3个至多24个字符）</label>
              <input
                type="text"
                value={newUser.username}
                onChange={(e) => setNewUser({ ...newUser, username: e.target.value })}
                placeholder="请输入用户名"
              />
            </div>
            <div className="form-group">
              <label>手机号</label>
              <input
                type="tel"
                value={newUser.phone}
                onChange={(e) => setNewUser({ ...newUser, phone: e.target.value })}
                placeholder="请输入手机号"
              />
            </div>
            <div className="form-group">
              <label>邮箱</label>
              <input
                type="email"
                value={newUser.email}
                onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
                placeholder="请输入邮箱"
              />
            </div>
            <div className="form-group">
              <label>角色</label>
              <select
                value={newUser.role}
                onChange={(e) => setNewUser({ ...newUser, role: e.target.value })}
              >
                <option value="watcher">观察者</option>
                <option value="admin">管理员</option>
              </select>
            </div>
            <div className="row">
              <button onClick={handleCreateUser}>确认创建</button>
              <button onClick={() => setCreateUserModal(false)}>取消</button>
            </div>
          </div>
        </div>
      )}

      {resetPasswordModal && (
        <div className="modal-mask" onClick={() => setResetPasswordModal(null)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <h2>重置密码</h2>
            <p>确定要重置用户 {resetPasswordModal.name} 的密码吗？</p>
            <p className="hint-text">系统将生成一个随机密码，只显示一次，请务必保存。</p>
            <div className="row">
              <button
                onClick={() => handleResetPassword(resetPasswordModal.id, resetPasswordModal.name)}
                className="error-btn"
              >
                确认重置
              </button>
              <button onClick={() => setResetPasswordModal(null)}>取消</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
