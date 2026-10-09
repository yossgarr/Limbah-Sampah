import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface WasteReport {
  id: string;
  category_code: string;
  input_amount: number;
  result_total: number;
  activity_date: string;
}

interface FormState {
  category_code: string;
  input_amount: string;
  activity_date: string;
}

interface FactorOption {
  category_code: string;
  name: string;
  value: number;
  unit: string;
}

const API_URL = 'http://localhost:8080/api/v1';
const API_BASE_URL = `${API_URL}/waste-reports`;

// Harus sama dengan batas di backend (report_handler.go)
const MAX_AMOUNT_KG = 1000;
const MAX_BACKDATE_DAYS = 365;

const todayString =() => new Date().toISOString().split('T')[0];

export const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [reports, setReports] = useState<WasteReport[]>([]);
  const [factors, setFactors] = useState<FactorOption[]>([]);
  const [loading, setLoading] = useState(true);

  // State Modal & Form
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<FormState>({
    category_code: '',
    input_amount: '',
    activity_date: todayString(),
  });
  const [formError, setFormError] = useState<string | null>(null);

  const userRole = localStorage.getItem('user_role') || 'user';
  const userEmail = localStorage.getItem('user_email') || 'User EcoLimbah';
  const canWrite = userRole !== 'observer';

  const handleLogout = () => {
    localStorage.clear();
    navigate('/login');
  };

  // fetch dengan token; sesi habis / tidak valid -> kembali ke login
  const authFetch = async (url: string, init: RequestInit = {}) => {
    const res = await fetch(url, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${localStorage.getItem('access_token')}`,
      },
    });
    if (res.status === 401) {
      handleLogout();
      throw new Error('Sesi berakhir, silakan login kembali');
    }
    return res;
  };

  const categoryName = (code: string) => factors.find((f) => f.category_code === code)?.name ?? code;

  // Fetch Data dari API Go Fiber
  const fetchReports = () => {
    setLoading(true);
    authFetch(API_BASE_URL)
      .then((res) => res.json())
      .then((data) => {
        setReports(Array.isArray(data) ? data : []);
        setLoading(false);
      })
      .catch((err) => {
        console.error('Gagal mengambil data:', err);
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchReports();
    authFetch(`${API_URL}/factors`)
      .then((res) => res.json())
      .then((data) => setFactors(Array.isArray(data) ? data : []))
      .catch((err) => console.error('Gagal mengambil kategori:', err));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Buka Modal Mode Tambah
  const handleOpenCreateModal = () => {
    setEditingId(null);
    setFormData({
      category_code: factors[0]?.category_code ?? '',
      input_amount: '',
      activity_date: todayString(),
    });
    setFormError(null);
    setIsModalOpen(true);
  };

  // Buka Modal Mode Edit
  const handleOpenEditModal = (report: WasteReport) => {
    setEditingId(report.id);
    setFormData({
      category_code: report.category_code,
      input_amount: report.input_amount.toString(),
      activity_date: report.activity_date.split('T')[0],
    });
    setFormError(null);
    setIsModalOpen(true);
  };

  // Submit Handler (Tambah / Update)
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    if (!formData.category_code) {
      setFormError('Kategori limbah wajib dipilih');
      return;
    }

    const amount = parseFloat(formData.input_amount);
    if (isNaN(amount) || amount <= 0) {
      setFormError('Jumlah sampah harus lebih dari 0 kg');
      return;
    }
    if (amount > MAX_AMOUNT_KG) {
      setFormError(`Jumlah sampah maksimal ${MAX_AMOUNT_KG} kg per laporan`);
      return;
    }

    const selectedDate = new Date(formData.activity_date);
    const today = new Date();
    today.setHours(23, 59, 59, 999);
    if (isNaN(selectedDate.getTime())) {
      setFormError('Tanggal wajib diisi');
      return;
    }
    if (selectedDate > today) {
      setFormError('Tanggal laporan tidak boleh di masa depan');
      return;
    }
    const earliest = new Date();
    earliest.setDate(earliest.getDate() - MAX_BACKDATE_DAYS);
    if (selectedDate < earliest) {
      setFormError(`Tanggal laporan tidak boleh lebih dari ${MAX_BACKDATE_DAYS} hari ke belakang`);
      return;
    }

    try {
      const isEdit = editingId !== null;
      const url = isEdit ? `${API_BASE_URL}/${editingId}` : API_BASE_URL;
      const method = isEdit ? 'PUT' : 'POST';

      const response = await authFetch(url, {
        method,
        body: JSON.stringify({
          category_code: formData.category_code,
          input_amount: amount,
          activity_date: formData.activity_date,
        }),
      });

      const resData = await response.json();
      if (!response.ok) throw new Error(resData.error || 'Gagal menyimpan data');

      setIsModalOpen(false);
      fetchReports();
    } catch (err: any) {
      setFormError(err.message);
    }
  };

  // Hapus Data Laporan
  const handleDelete = async (id: string) => {
    if (!window.confirm('Apakah Anda yakin ingin menghapus laporan ini?')) return;

    try {
      const response = await authFetch(`${API_BASE_URL}/${id}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const resData = await response.json();
        throw new Error(resData.error || 'Gagal menghapus laporan');
      }

      fetchReports();
    } catch (err: any) {
      alert(err.message);
    }
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
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
          <h2>Riwayat Laporan Limbah (Realtime Neon DB)</h2>
          {canWrite && (
            <button
              onClick={handleOpenCreateModal}
              disabled={factors.length === 0}
              style={{ padding: '8px 16px', backgroundColor: '#059669', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold' }}
            >
              + Tambah Laporan
            </button>
          )}
        </div>

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
                {canWrite && <th style={{ padding: '8px', border: '1px solid #ddd', textAlign: 'center' }}>Aksi</th>}
              </tr>
            </thead>
            <tbody>
              {reports.length === 0 ? (
                <tr>
                  <td colSpan={canWrite ? 5 : 4} style={{ padding: '16px', textAlign: 'center', color: '#6b7280' }}>
                    Belum ada laporan limbah.
                  </td>
                </tr>
              ) : (
                reports.map((item) => (
                  <tr key={item.id}>
                    <td style={{ padding: '8px', border: '1px solid #ddd' }}>{categoryName(item.category_code)}</td>
                    <td style={{ padding: '8px', border: '1px solid #ddd' }}>{item.input_amount}</td>
                    <td style={{ padding: '8px', border: '1px solid #ddd' }}>{item.result_total}</td>
                    <td style={{ padding: '8px', border: '1px solid #ddd' }}>{item.activity_date.split('T')[0]}</td>
                    {canWrite && (
                      <td style={{ padding: '8px', border: '1px solid #ddd', textAlign: 'center' }}>
                        <button
                          onClick={() => handleOpenEditModal(item)}
                          style={{ padding: '4px 8px', marginRight: '8px', backgroundColor: '#d97706', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDelete(item.id)}
                          style={{ padding: '4px 8px', backgroundColor: '#dc2626', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                        >
                          Hapus
                        </button>
                      </td>
                    )}
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}
      </div>

      {/* Modal Form Tambah/Edit */}
      {isModalOpen && (
        <div style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)', display: 'flex', justifyContent: 'center', alignItems: 'center', zIndex: 1000
        }}>
          <div style={{ backgroundColor: 'white', padding: '24px', borderRadius: '8px', width: '100%', maxWidth: '400px' }}>
            <h2 style={{ marginTop: 0 }}>{editingId ? 'Edit Laporan Limbah' : 'Tambah Laporan Limbah'}</h2>
            
            {formError && (
              <div style={{ backgroundColor: '#fee2e2', color: '#dc2626', padding: '8px', borderRadius: '4px', marginBottom: '16px', fontSize: '14px' }}>
                {formError}
              </div>
            )}

            <form onSubmit={handleSubmit}>
              <div style={{ marginBottom: '16px' }}>
                <label style={{ display: 'block', marginBottom: '4px', fontWeight: 'bold' }}>Kategori Limbah</label>
                <select
                  value={formData.category_code}
                  onChange={(e) => setFormData({ ...formData, category_code: e.target.value })}
                  style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
                >
                  {factors.map((f) => (
                    <option key={f.category_code} value={f.category_code}>
                      {f.name} ({f.category_code})
                    </option>
                  ))}
                </select>
              </div>

              <div style={{ marginBottom: '16px' }}>
                <label style={{ display: 'block', marginBottom: '4px', fontWeight: 'bold' }}>Jumlah Input (Kg)</label>
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  max={MAX_AMOUNT_KG}
                  placeholder="Contoh: 5.5"
                  value={formData.input_amount}
                  onChange={(e) => setFormData({ ...formData, input_amount: e.target.value })}
                  style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc', boxSizing: 'border-box' }}
                  required
                />
              </div>

              <div style={{ marginBottom: '20px' }}>
                <label style={{ display: 'block', marginBottom: '4px', fontWeight: 'bold' }}>Tanggal Aktivitas</label>
                <input
                  type="date"
                  value={formData.activity_date}
                  onChange={(e) => setFormData({ ...formData, activity_date: e.target.value })}
                  style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc', boxSizing: 'border-box' }}
                  required
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  style={{ padding: '8px 16px', backgroundColor: '#6b7280', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                >
                  Batal
                </button>
                <button
                  type="submit"
                  style={{ padding: '8px 16px', backgroundColor: '#059669', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                >
                  {editingId ? 'Simpan Perubahan' : 'Tambah Laporan'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};