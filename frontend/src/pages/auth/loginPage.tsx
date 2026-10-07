import React, { useState } from 'react';

export const LoginPage = () => {
  const [showExternal, setShowExternal] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  const handleSSO = () => {
    // Arahkan ke endpoint authorization Keycloak (FR-AUTH-1)[cite: 1]
    window.location.href = 'http://localhost:8080/api/v1/auth/sso/login';
  };

  const handleExternalSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch('http://localhost:8080/api/v1/auth/external/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });
      
      const data = await res.json();
      if (!res.ok) {
        // Pesan error seragam: "Email atau password salah" (FR-AUTH-11)[cite: 1]
        setErrorMsg(data.error);
        return;
      }
      
      // Simpan token ke state/localStorage
      console.log('Login eksternal berhasil sebagai observer', data);
    } catch (err) {
      setErrorMsg('Terjadi kesalahan pada server.');
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-50">
      <div className="w-full max-w-md p-8 bg-white rounded shadow">
        <h1 className="mb-6 text-2xl font-bold text-center text-green-700">Limbah-Sampah</h1>
        
        <button 
          onClick={handleSSO}
          className="w-full py-3 mb-4 text-white bg-green-600 rounded hover:bg-green-700"
        >
          Masuk dengan SSO Pegawai
        </button>

        {!showExternal ? (
          <div className="text-center">
            <button 
              onClick={() => setShowExternal(true)}
              className="text-sm text-gray-500 underline hover:text-gray-700"
            >
              Masuk sebagai pengamat eksternal
            </button>
          </div>
        ) : (
          <form onSubmit={handleExternalSubmit} className="pt-4 border-t border-gray-200">
            <h2 className="mb-4 text-sm font-semibold text-gray-700">Login Pengamat Eksternal</h2>
            {errorMsg && <p className="mb-4 text-sm text-red-500">{errorMsg}</p>}
            
            <div className="mb-4">
              <label className="block mb-1 text-sm text-gray-600">Email</label>
              <input 
                type="email" 
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded"
                required
              />
            </div>
            
            <div className="mb-6">
              <label className="block mb-1 text-sm text-gray-600">Kata Sandi</label>
              <input 
                type="password" 
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded"
                required
              />
            </div>
            
            <button 
              type="submit"
              className="w-full py-2 text-white bg-gray-800 rounded hover:bg-gray-900"
            >
              Masuk
            </button>
          </form>
        )}
      </div>
    </div>
  );
};