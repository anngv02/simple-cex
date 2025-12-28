import { useEffect, useState } from 'react';
import { WS_URL } from '../config';

// Định nghĩa kiểu dữ liệu cho Lệnh
interface Order {
  ID: number;
  Price: number;
  Amount: number;
}

interface OrderBookData {
  bids: Order[]; // Người mua (Xanh)
  asks: Order[]; // Người bán (Đỏ)
}

export default function OrderBook() {
  const [book, setBook] = useState<OrderBookData>({ bids: [], asks: [] });

  useEffect(() => {
    // 1. Kết nối WebSocket tới Backend
    const ws = new WebSocket(`${WS_URL}/ws`);
    let isConnected = false;

    ws.onopen = () => {
      console.log("Connected to WebSocket");
      isConnected = true;
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected");
      isConnected = false;
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        // Chỉ cập nhật khi message type là ORDERBOOK_UPDATE
        if (data.type === "ORDERBOOK_UPDATE") {
          // Giới hạn chỉ hiển thị 10 orders cho mỗi bên
          const bids = (data.bids || []).slice(0, 10);
          const asks = (data.asks || []).slice(0, 10);
          
          setBook({
              bids: bids,
              asks: asks
          });
        }
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };

    return () => {
      // Chỉ đóng nếu đã kết nối
      if (isConnected && ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }, []);

  return (
    <div className="p-4 bg-gray-900 rounded-lg w-full max-w-md">
      <h2 className="text-xl font-bold mb-4 text-white">Order Book (BTC/USDT)</h2>
      
      <div className="flex justify-between text-xs text-gray-400 mb-2 font-mono">
        <span>Price (USDT)</span>
        <span>Amount (BTC)</span>
      </div>

      {/* ASKS (Người bán - Màu Đỏ) - Xếp ngược từ cao xuống thấp để giá thấp nhất ở gần giữa */}
      <div className="flex flex-col-reverse mb-2"> 
        {book.asks.map((ask) => (
          <div key={ask.ID} className="flex justify-between text-red-500 font-mono hover:bg-gray-800 cursor-pointer">
            <span>{ask.Price.toLocaleString()}</span>
            <span>{ask.Amount.toFixed(4)}</span>
          </div>
        ))}
      </div>

      {/* Giá hiện tại (Current Price) - Để trống hoặc giả lập */}
      <div className="py-2 text-center text-xl font-bold text-white border-y border-gray-700 my-2">
         {book.bids.length > 0 ? book.bids[0].Price.toLocaleString() : "---"}
      </div>

      {/* BIDS (Người mua - Màu Xanh) */}
      <div>
        {book.bids.map((bid) => (
          <div key={bid.ID} className="flex justify-between text-green-500 font-mono hover:bg-gray-800 cursor-pointer">
            <span>{bid.Price.toLocaleString()}</span>
            <span>{bid.Amount.toFixed(4)}</span>
          </div>
        ))}
      </div>
    </div>
  );
}