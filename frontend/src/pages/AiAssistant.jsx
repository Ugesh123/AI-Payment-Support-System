/**
 * AI Assistant chat page.
 * User picks a transaction, then asks one of a few suggested
 * questions (or types their own). Chat history is session-only.
 */
import { useEffect, useState, useRef } from "react";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Loader from "../components/ui/Loader";
import ErrorMessage from "../components/ui/ErrorMessage";
import { getTransactions } from "../services/transactionService";
import { askAssistant } from "../services/assistantService";

const SUGGESTED_QUESTIONS = [
  "Why did my payment fail?",
  "What is my payment status?",
  "Where is my refund?",
  "What should I do next?",
  "Show my payment details.",
];

export default function AiAssistant() {
  const [transactions, setTransactions] = useState([]);
  const [selectedTxnId, setSelectedTxnId] = useState("");
  const [question, setQuestion] = useState("");
  const [messages, setMessages] = useState([]);
  const [asking, setAsking] = useState(false);
  const [loadError, setLoadError] = useState("");
  const [askError, setAskError] = useState("");
  const chatEndRef = useRef(null);

  useEffect(() => {
    const loadTransactions = async () => {
      try {
        const data = await getTransactions();
        setTransactions(data);
        if (data.length > 0) setSelectedTxnId(data[0].transaction_id);
      } catch (err) {
        setLoadError(err.response?.data?.detail || "Failed to load transactions.");
      }
    };
    loadTransactions();
  }, []);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleAsk = async (questionText) => {
    if (!selectedTxnId || !questionText.trim()) return;

    setAskError("");
    setMessages((prev) => [...prev, { role: "user", text: questionText }]);
    setQuestion("");
    setAsking(true);

    try {
      const result = await askAssistant(selectedTxnId, questionText);
      setMessages((prev) => [...prev, { role: "ai", text: result.answer }]);
    } catch (err) {
      setAskError(err.response?.data?.detail || "The assistant is unavailable right now.");
    } finally {
      setAsking(false);
    }
  };

  const handleFormSubmit = (e) => {
    e.preventDefault();
    handleAsk(question);
  };

  if (loadError) return <ErrorMessage message={loadError} />;

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-800 mb-5">AI Payment Assistant</h1>

      {transactions.length === 0 ? (
        <Card>
          <p className="text-gray-500 text-sm">
            Make a payment first so you have a transaction to ask about.
          </p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
          <Card className="lg:col-span-1 h-fit">
            <label className="block text-sm text-gray-600 mb-1">Select Transaction</label>
            <select
              value={selectedTxnId}
              onChange={(e) => setSelectedTxnId(e.target.value)}
              className="w-full border rounded px-3 py-2 text-sm mb-4 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {transactions.map((txn) => (
                <option key={txn.transaction_id} value={txn.transaction_id}>
                  {txn.transaction_id} -- {txn.status}
                </option>
              ))}
            </select>

            <p className="text-sm text-gray-600 mb-2">Suggested Questions</p>
            <div className="flex flex-col gap-2">
              {SUGGESTED_QUESTIONS.map((q) => (
                <button
                  key={q}
                  onClick={() => handleAsk(q)}
                  disabled={asking}
                  className="text-left text-sm border rounded px-3 py-2 text-gray-700 hover:bg-blue-50 hover:border-blue-300 transition disabled:opacity-50"
                >
                  {q}
                </button>
              ))}
            </div>
          </Card>

          <Card className="lg:col-span-2 flex flex-col min-h-[400px] max-h-[70vh] lg:h-[520px]">
            <div className="flex-1 overflow-y-auto space-y-3 pr-1">
              {messages.length === 0 && (
                <p className="text-sm text-gray-400 text-center mt-10">
                  Ask a question about the selected transaction to get started.
                </p>
              )}

              {messages.map((msg, i) => (
                <ChatBubble key={i} role={msg.role} text={msg.text} />
              ))}

              {asking && <Loader label="Assistant is thinking..." />}
              {askError && <ErrorMessage message={askError} />}

              <div ref={chatEndRef} />
            </div>

            <form onSubmit={handleFormSubmit} className="flex gap-2 mt-4 pt-4 border-t">
              <input
                type="text"
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                placeholder="Type your own question..."
                className="flex-1 border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <Button type="submit" loading={asking}>Ask</Button>
            </form>
          </Card>
        </div>
      )}
    </div>
  );
}

function ChatBubble({ role, text }) {
  const isUser = role === "user";
  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"}`}>
      <div
        className={`max-w-[80%] rounded-lg px-4 py-2 text-sm whitespace-pre-line ${
          isUser ? "bg-blue-600 text-white" : "bg-gray-100 text-gray-800"
        }`}
      >
        {text}
      </div>
    </div>
  );
}
