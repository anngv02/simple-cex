import { useEffect, useRef, useState } from 'react';
import { createChart, CandlestickSeries } from 'lightweight-charts';
import axios from 'axios';
import { API_URL, WS_URL } from '../config';

interface OHLCV {
  time: number; // Unix timestamp
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

// Định nghĩa interface cho candlestick data
interface CandlestickData {
  time: number | string; // Time có thể là number (Unix timestamp) hoặc string
  open: number;
  high: number;
  low: number;
  close: number;
}

// Hàm cập nhật chart khi có trade mới
const updateChartWithTrade = (
  candlestickSeries: any,
  tradePrice: number,
  tradeTime: number,
  currentInterval: string = '1m'
) => {
  if (!candlestickSeries) return;

  // Tính timestamp của nến (làm tròn xuống theo interval)
  const tradeTimeSeconds = Math.floor(tradeTime / 1000);
  let intervalSeconds = 60; // 1m default
  
  switch (currentInterval) {
    case '1m':
      intervalSeconds = 60;
      break;
    case '5m':
      intervalSeconds = 300;
      break;
    case '15m':
      intervalSeconds = 900;
      break;
    case '1h':
      intervalSeconds = 3600;
      break;
  }
  
  const candleTime = Math.floor(tradeTimeSeconds / intervalSeconds) * intervalSeconds;

  // Lấy dữ liệu hiện tại (nếu có method data())
  let existingData: any[] = [];
  if (typeof candlestickSeries.data === 'function') {
    existingData = candlestickSeries.data();
  }
  
  // Tìm nến hiện tại trong data
  let currentCandle: any = null;
  for (let i = existingData.length - 1; i >= 0; i--) {
    const candle = existingData[i];
    if (candle && candle.time === candleTime) {
      currentCandle = candle;
      break;
    }
  }

  if (currentCandle) {
    // Cập nhật nến hiện có
    const updatedCandle: CandlestickData = {
      time: candleTime,
      open: currentCandle.open,
      high: Math.max(currentCandle.high, tradePrice),
      low: Math.min(currentCandle.low, tradePrice),
      close: tradePrice,
    };
    candlestickSeries.update(updatedCandle);
  } else {
    // Tạo nến mới
    const newCandle: CandlestickData = {
      time: candleTime,
      open: tradePrice,
      high: tradePrice,
      low: tradePrice,
      close: tradePrice,
    };
    candlestickSeries.update(newCandle);
  }
};

export default function CandlestickChart() {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<ReturnType<typeof createChart> | null>(null);
  const candlestickSeriesRef = useRef<any>(null); // Type sẽ được infer từ giá trị
  const [isLoading, setIsLoading] = useState(true);
  const [interval, setInterval] = useState<string>('1m'); // 1m, 5m, 15m, 1h
  const wsRef = useRef<WebSocket | null>(null);
  const intervalRef = useRef<number | null>(null);

  useEffect(() => {
    if (!chartContainerRef.current) return;

    // Đợi một chút để đảm bảo container đã có kích thước
    const initChart = () => {
      if (!chartContainerRef.current) return;

      // Đảm bảo container có width
      let containerWidth = chartContainerRef.current.clientWidth;
      if (containerWidth === 0) {
        // Nếu chưa có width, thử lấy từ parent hoặc dùng default
        containerWidth = chartContainerRef.current.parentElement?.clientWidth || 800;
      }

      // Tạo chart
      const chart = createChart(chartContainerRef.current, {
        layout: {
          background: { color: '#0b0e11' },
          textColor: '#eaecef',
        },
        grid: {
          vertLines: { color: '#1a1d29' },
          horzLines: { color: '#1a1d29' },
        },
        width: containerWidth,
        height: 500,
        timeScale: {
          timeVisible: true,
          secondsVisible: false,
        },
      });

      // Tạo candlestick series
      // Với lightweight-charts v5, dùng addSeries với CandlestickSeries
      const createSeries = (): any => {
        const chartAny = chart as any;
        
        // Với v5, dùng addSeries với CandlestickSeries
        if (typeof chartAny.addSeries === 'function' && CandlestickSeries) {
          return chartAny.addSeries(CandlestickSeries, {
            upColor: '#26a69a',
            downColor: '#ef5350',
            borderVisible: false,
            wickUpColor: '#26a69a',
            wickDownColor: '#ef5350',
          });
        }
        
        throw new Error('Cannot create candlestick series - addSeries or CandlestickSeries not available');
      };
      
      // Tạo series và sau đó fetch data
      Promise.resolve(createSeries())
        .then((candlestickSeries: any) => {
          if (!candlestickSeries) {
            throw new Error('Failed to create candlestick series');
          }
          
          chartRef.current = chart;
          candlestickSeriesRef.current = candlestickSeries;
          
          // Fetch dữ liệu từ API với interval hiện tại
          const fetchData = async (currentInterval: string = interval) => {
            try {
              const response = await axios.get(`${API_URL}/trades/BTC_USDT?interval=${currentInterval}&limit=100`);
              const trades: OHLCV[] = response.data;

              if (trades && trades.length > 0) {
                // Chuyển đổi sang format của lightweight-charts
                const chartData: CandlestickData[] = trades.map((trade) => ({
                  time: trade.time / 1000, // Chuyển từ milliseconds sang seconds
                  open: trade.open,
                  high: trade.high,
                  low: trade.low,
                  close: trade.close,
                }));

                candlestickSeries.setData(chartData);
                chart.timeScale().fitContent();
              } else {
                // Nếu không có dữ liệu, hiển thị chart trống
                candlestickSeries.setData([]);
                console.log('No trades data available. Waiting for simulation to start...');
              }
              setIsLoading(false);
            } catch (error) {
              console.error('Error fetching chart data:', error);
              // Nếu API lỗi, hiển thị chart trống
              candlestickSeries.setData([]);
              setIsLoading(false);
            }
          };

          // Fetch dữ liệu ban đầu
          fetchData(interval);

          // 4. Tự động cập nhật chart mỗi 1 giây
          intervalRef.current = window.setInterval(() => {
            fetchData(interval);
          }, 1000); // 1000ms = 1 giây

          // 5. Kết nối WebSocket để nhận trade updates real-time
          const ws = new WebSocket(`${WS_URL}/ws`);
          wsRef.current = ws;

          ws.onopen = () => {
            console.log("Chart WebSocket connected");
          };

          ws.onmessage = (event) => {
            try {
              const data = JSON.parse(event.data);
              
              // Xử lý TRADE_UPDATE để cập nhật chart (chỉ update real-time khi interval là 1m)
              if (data.type === "TRADE_UPDATE" && data.symbol === "BTC_USDT" && candlestickSeries && interval === '1m') {
                updateChartWithTrade(candlestickSeries, data.price, data.time, interval);
              }
            } catch (error) {
              console.error("Error parsing WebSocket message:", error);
            }
          };

          ws.onerror = (error) => {
            console.error("Chart WebSocket error:", error);
          };

          ws.onclose = () => {
            console.log("Chart WebSocket disconnected");
          };
        })
        .catch((error: any) => {
          console.error('Failed to create series:', error);
          setIsLoading(false);
        });
      
      // Return early, sẽ tiếp tục trong promise
      return;

      // Handle resize
      const handleResize = () => {
        if (chartContainerRef.current && chartRef.current) {
          const newWidth = chartContainerRef.current.clientWidth;
          if (newWidth > 0) {
            chartRef.current.applyOptions({
              width: newWidth,
            });
          }
        }
      };

      window.addEventListener('resize', handleResize);

      // Trả về cleanup function
      return () => {
        window.removeEventListener('resize', handleResize);
      };
    };

    // Sử dụng setTimeout để đảm bảo DOM đã render xong
    let cleanupFn: (() => void) | null = null;
    const timer = setTimeout(() => {
      cleanupFn = initChart() || null;
    }, 50);

    return () => {
      clearTimeout(timer);
      if (cleanupFn) {
        cleanupFn();
      }
      // Clear interval
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      // Đóng WebSocket
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      if (chartRef.current) {
        chartRef.current.remove();
        chartRef.current = null;
      }
    };
  }, [interval]); // Re-run khi interval thay đổi

  return (
    <div className="w-full bg-gray-900 rounded-lg p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold text-white">BTC/USDT Chart</h2>
        <div className="flex gap-2">
          <button 
            onClick={() => setInterval('1m')}
            className={`px-3 py-1 rounded text-sm ${
              interval === '1m' 
                ? 'bg-yellow-500 text-white' 
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            1m
          </button>
          <button 
            onClick={() => setInterval('5m')}
            className={`px-3 py-1 rounded text-sm ${
              interval === '5m' 
                ? 'bg-yellow-500 text-white' 
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            5m
          </button>
          <button 
            onClick={() => setInterval('15m')}
            className={`px-3 py-1 rounded text-sm ${
              interval === '15m' 
                ? 'bg-yellow-500 text-white' 
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            15m
          </button>
          <button 
            onClick={() => setInterval('1h')}
            className={`px-3 py-1 rounded text-sm ${
              interval === '1h' 
                ? 'bg-yellow-500 text-white' 
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            1h
          </button>
        </div>
      </div>
      {isLoading && (
        <div className="flex items-center justify-center h-[500px] text-gray-400">
          Loading chart...
        </div>
      )}
      <div 
        ref={chartContainerRef} 
        className="w-full" 
        style={{ 
          height: '500px',
          minHeight: '500px',
          width: '100%'
        }} 
      />
    </div>
  );
}

