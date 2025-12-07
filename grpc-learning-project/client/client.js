const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const path = require('path');

// Defined in our proto file
const PROTO_PATH = path.join(__dirname, '../protos/user.proto');

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});

// Load the package definition into gRPC
const userProto = grpc.loadPackageDefinition(packageDefinition).user;

const client = new userProto.UserService(
    'localhost:50051',
    grpc.credentials.createInsecure()
);

// Helper to run client
function main() {
    console.log("Client starting...");

    // 1. Create a User
    const newUser = {
        name: "Reagan Test",
        email: "reagan@example.com",
        age: 30
    };

    console.log("Creating user...", newUser);

    client.CreateUser(newUser, (err, response) => {
        if (err) {
            console.error("Error creating user:", err);
            return;
        }
        console.log("✅ User created successfully:", response);

        const createdId = response.id;

        // 2. Get the User back
        console.log(`Fetching user with ID: ${createdId}...`);

        client.GetUser({ id: createdId }, (err, response) => {
            if (err) {
                console.error("Error fetching user:", err);
                return;
            }
            console.log("✅ User fetched successfully:", response);
        });
    });
}

main();
