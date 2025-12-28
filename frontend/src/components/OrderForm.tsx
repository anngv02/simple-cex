import { useState } from 'react';
import axios from 'axios';

export default function OrderForm() {
  const [side, setSide] = useState<'BUY' | 'SELL'>('BUY');
  const [price, setPrice] = useState('');
  const [amount, setAmount] = useState('');
  const [userID, setUserID] = useState('1'); // Giả lập User ID (sau này sẽ lấy từ Login)

  const handleSubmit = async () => {
    try {
      // Gọi API Backend
      await axios.post('http://localhost:8010/order', {
        user_id: parseInt(userID),
        symbol: "BTC_USDT",
        side: side,
        price: parseFloat(price),
        amount: parseFloat(amount)
      });
      alert("Đặt lệnh thành công!");
      // Reset form (tuỳ chọn)
    } catch (error) {
      console.error(error);
      alert("Lỗi đặt lệnh");
    }
  };

  return (
    <div className="p-4 bg-gray-900 rounded-lg w-full max-w-md h-fit">
      {/* Tab Mua/Bán */}
      <div className="flex mb-4 bg-gray-800 rounded p-1">
        <button 
          onClick={() => setSide('BUY')}
          className={`flex-1 py-2 rounded font-bold ${side === 'BUY' ? 'bg-green-600 text-white' : 'text-gray-400'}`}
        >
          BUY
        </button>
        <button 
          onClick={() => setSide('SELL')}
          className={`flex-1 py-2 rounded font-bold ${side === 'SELL' ? 'bg-red-600 text-white' : 'text-gray-400'}`}
        >
          SELL
        </button>
      </div>

      {/* Input Form */}
      <div className="space-y-3">
        <div>
            <label className="text-xs text-gray-400">User ID (Simulate)</label>
            <input 
                type="number" 
                className="w-full bg-gray-800 text-white p-2 rounded outline-none border border-gray-700 focus:border-yellow-500"
                value={userID}
                onChange={e => setUserID(e.target.value)}
            />
        </div>
        <div>
            <label className="text-xs text-gray-400">Price (USDT)</label>
            <input 
                type="number" 
                className="w-full bg-gray-800 text-white p-2 rounded outline-none border border-gray-700 focus:border-yellow-500"
                value={price}
                onChange={e => setPrice(e.target.value)}
                placeholder="50000"
            />
        </div>
        <div>
            <label className="text-xs text-gray-400">Amount (BTC)</label>
            <input 
                type="number" 
                className="w-full bg-gray-800 text-white p-2 rounded outline-none border border-gray-700 focus:border-yellow-500"
                value={amount}
                onChange={e => setAmount(e.target.value)}
                placeholder="0.1"
            />
        </div>
        
        <button 
            onClick={handleSubmit}
            className={`w-full py-3 rounded font-bold mt-4 ${side === 'BUY' ? 'bg-green-600 hover:bg-green-500' : 'bg-red-600 hover:bg-red-500'}`}
        >
            {side} BTC
        </button>
      </div>
    </div>
  );
}