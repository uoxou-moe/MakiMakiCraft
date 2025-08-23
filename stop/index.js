exports.handler = async (event) => {
    console.log("Hello from Node.js Lambda!");
    console.log("Received event:", JSON.stringify(event, null, 2));

    const response = {
        statusCode: 200,
        body: JSON.stringify('Hello from a Node.js Lambda function managed by Terraform!'),
    };
    return response;
};
