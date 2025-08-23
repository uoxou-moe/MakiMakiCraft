exports.handler = async (event) => {
    console.log("init: Lambda function executed successfully!");

    const response = {
        statusCode: 200,
        body: JSON.stringify('Initial function is working.'),
    };
    return response;
};
