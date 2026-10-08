import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

export const Login: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();

  const handleExternalLogin = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const response = await fetch('http://localhost:8080/api/v1/auth/login-external', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      const data = await response.json();

      if (response.ok) {
        // Simpan sesi login
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('user_role', data.role);
        localStorage.setItem('user_email', email);

        // Redirect langsung ke Dashboard
        navigate('/dashboard');
      } else {
        alert(data.error || 'Login gagal');
      }
    } catch (err) {
      console.error(err);
      alert('Terjadi kesalahan koneksi ke server');
    }
  };

  return (
    <div style={{ maxWidth: '400px', margin: '50px auto', padding: '20px', border: '1px solid #ccc' }}>
      <h2>Login EcoLimbah</h2>
      <form onSubmit={handleExternalLogin}>
        <div style={{ marginBottom: '12px' }}>
          <label>Email</label>
          <input 
            type="email" 
            value={email} 
            onChange={(e) => setEmail(e.target.value)} 
            required 
            style={{ width: '100%', padding: '8px', marginTop: '4px' }}
          />
        </div>
        <div style={{ marginBottom: '16px' }}>
          <label>Password</label>
          <input 
            type="password" 
            value={password} 
            onChange={(e) => setPassword(e.target.value)} 
            required 
            style={{ width: '100%', padding: '8px', marginTop: '4px' }}
          />
        </div>
        <button type="submit" style={{ width: '100%', padding: '10px', backgroundColor: '#16a34a', color: 'white', border: 'none' }}>
          Login External
        </button>
      </form>
    </div>
  );
};