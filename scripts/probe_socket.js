const io = require('socket.io-client');

const socket = io('https://quotes.instantswap.app', {
    transports: ['websocket', 'polling'],
    reconnection: true,
});

socket.on('connect', () => {
    console.log('Connected to Server');
});

socket.on('disconnect', () => {
    console.log('Disconnected');
});

socket.on('connect_error', (err) => {
    console.log('Connection Error:', err.message);
});

// Try to subscribe
socket.on('connect', () => {
    console.log('Connected to Server - Attempting subscriptions...');

    // Guess 1: Common crypto patterns
    socket.emit('subscribe', ['BTC', 'ETH', 'USDT']);
    socket.emit('subscribe', 'rates');
    socket.emit('subscribe', { currencyFrom: 'BTC', currencyTo: 'ETH' });

    // Guess 2: "join" room
    socket.emit('join', 'rates');
    socket.emit('join', 'prices');

    // Guess 3: "get_rates"
    socket.emit('get_rates');
});

// Log any event
const onevent = socket.onevent;
socket.onevent = function (packet) {
    const args = packet.data || [];
    onevent.call(this, packet);    // original call
    packet.data = ["*"].concat(args);
    onevent.call(this, packet);      // additional call to catch-all
};

// Rate limit logging
let logCount = 0;

socket.on("*", function (event, data) {
    if (logCount < 20) {
        console.log("=====================================");
        console.log("EVENT:", event);
        if (Array.isArray(data)) {
            console.log("DATA IS ARRAY. Length:", data.length);
            if (data.length > 0) console.log("SAMPLE:", JSON.stringify(data[0]).substring(0, 100));
        } else if (typeof data === 'object') {
            console.log("DATA TYPE:", typeof data);
            const preview = JSON.stringify(data).substring(0, 200);
            console.log("PREVIEW:", preview);
        } else {
            console.log("DATA:", data);
        }
        logCount++;
    }
});

// Keep alive
setInterval(() => {
    console.log('Waiting for data...');
    // Periodic ping or re-subscribe
    socket.emit('get_rates');
}, 5000);
