import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface WasteReport {
  id: string;
  category_code: string;
  input_amount: number;
  result_total: number;
  activity_date: string;
}

export const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [reports, setReports] = useState<WasteReport[]>([]);
  const [loading, setLoading] = useState(true);

  const userRole = localStorage.getItem('user_role') || 'user';
  const userEmail = localStorage.getItem('user_email') || 'User EcoLimbah';

  useEffect(() => {
    // Ambil data dari API Go Fiber
    fetch('http://localhost:8080/api/v1/waste-reports')
      .then((res) => res.json())
      .then((data) => {
        setReports(data || []);
        setLoading(false);
      })
      .catch((err) => {
        console.error('Gagal mengambil data:', err);
        setLoading(false);
      });
  }, []);

  const handleLogout = () => {
    localStorage.clear();
    navigate('/login');
  };

  // Kalkulasi otomatis dari data realtime database
  const totalSampah = reports.reduce((acc, item) => acc + item.input_amount, 0);
  const totalPoin = reports.reduce((acc, item) => acc + item.result_total, 0);

  return (
    <div style={{ padding: '24px', fontFamily: 'sans-serif' }}>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '32px' }}>
        <div>
          <h1>Dashboard EcoLimbah</h1>
          <p>Selamat datang, <strong>{userEmail}</strong> ({userRole.toUpperCase()})</p>
        </div>
        <button 
          onClick={handleLogout}
          style={{ padding: '8px 16px', backgroundColor: '#dc2626', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
        >
          Logout
        </button>
      </header>

      {/* Ringkasan Stat Card Dinamis */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', marginBottom: '32px' }}>
        <div style={{ border: '1px solid #e5e7eb', padding: '16px', borderRadius: '8px' }}>
          <h3>Total Laporan</h3>
          <p style={{ fontSize: '24px', fontWeight: 'bold' }}>{loading ? '...' : reports.length}</p>
        </div>
        <div style={{ border: '1px solid #e5e7eb', padding: '16px', borderRadius: '8px' }}>
          <h3>Total Sampah</h3>
          <p style={{ fontSize: '24px', fontWeight: 'bold' }}>{loading ? '...' : `${totalSampah} Kg`}</p>
        </div>
        <div style={{ border: '1px solid #e5e7eb', padding: '16px', borderRadius: '8px' }}>
          <h3>Total Poin</h3>
          <p style={{ fontSize: '24px', fontWeight: 'bold' }}>{loading ? '...' : totalPoin}</p>
        </div>
      </div>

      {/* Tabel Data Realtime */}
      <div style={{ border: '1px solid #e5e7eb', padding: '24px', borderRadius: '8px' }}>
        <h2>Riwayat Laporan Limbah (Realtime Neon DB)</h2>
        {loading ? (
          <p>Memuat data...</p>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', marginTop: '16px' }}>
            <thead>
              <tr style={{ backgroundColor: '#f3f4f6', textAlign: 'left' }}>
                <th style={{ padding: '8px', border: '1px solid #ddd' }}>Kategori</th>
                <th style={{ padding: '8px', border: '1px solid #ddd' }}>Jumlah (Kg)</th>
                <th style={{ padding: '8px', border: '1px solid #ddd' }}>Poin Ditambah</th>
                <th style={{ padding: '8px', border: '1px solid #ddd' }}>Tanggal</th>
              </tr>
            </thead>
            <tbody>
              {reports.map((item) => (
                <tr key={item.id}>
                  <td style={{ padding: '8px', border: '1px solid #ddd' }}>{item.category_code}</td>
                  <td style={{ padding: '8px', border: '1px solid #ddd' }}>{item.input_amount}</td>
                  <td style={{ padding: '8px', border: '1px solid #ddd' }}>{item.result_total}</td>
                  <td style={{ padding: '8px', border: '1px solid #ddd' }}>{item.activity_date.split('T')[0]}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};