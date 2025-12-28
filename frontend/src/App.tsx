import OrderBook from './components/Orderbook';
import OrderForm from './components/OrderForm';
import CandlestickChart from './components/CandlestickChart';

function App() {
  return (
    <div className="min-h-screen p-4 bg-gray-950">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold mb-6 text-yellow-500 text-center">
          SIMPLE CEX <span className="text-sm text-gray-400">v1.0</span>
        </h1>
        
        {/* Chart nến - Chiếm toàn bộ chiều rộng */}
        <div className="mb-6">
          <CandlestickChart />
        </div>

        {/* Phần dưới: Orderbook và Form đặt lệnh */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Cột trái: Orderbook */}
          <div>
            <OrderBook />
          </div>

          {/* Cột phải: Form đặt lệnh */}
          <div>
            <OrderForm />
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;