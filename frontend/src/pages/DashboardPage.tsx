import type { ComponentType, ReactNode } from 'react';
import { formatEuro } from '@/lib/format';
import { useGetDashboardStatsQuery, useGetRecentOrdersQuery } from '@/store/api/appApi';
import { logger } from '@/lib/logger';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { ClipboardList, Truck, Users, FileText, TrendingUp } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import { StatusBadge } from '@/components/shared/StatusBadge';

interface KPICardProps {
  title: string;
  value: ReactNode;
  icon: ComponentType<{ className?: string }>;
  description?: string;
  testId?: string;
}

const KPICard = ({ title, value, icon: Icon, description, testId }: KPICardProps) => (
  <Card data-testid={testId || 'kpi-card'} className="shadow-sm">
    <CardContent className="p-4 lg:p-5">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs font-medium text-muted-foreground mb-1">{title}</p>
          <p className="text-2xl md:text-3xl font-bold tracking-tight" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>{value}</p>
          {description && <p className="text-xs text-muted-foreground mt-1">{description}</p>}
        </div>
        <div className="p-2 rounded-lg bg-accent">
          <Icon className="h-4 w-4 text-primary" />
        </div>
      </div>
    </CardContent>
  </Card>
);

export default function DashboardPage() {
  const statsQuery = useGetDashboardStatsQuery();
  const ordersQuery = useGetRecentOrdersQuery();

  const stats = statsQuery.data;
  const orders = ordersQuery.data ?? [];
  const loading = statsQuery.isLoading || ordersQuery.isLoading;

  if (statsQuery.error) logger.error('Dashboard stats error:', statsQuery.error);
  if (ordersQuery.error) logger.error('Dashboard orders error:', ordersQuery.error);

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4">
          {[1, 2, 3, 4].map(i => <Skeleton key={i} className="h-24 rounded-xl" />)}
        </div>
        <Skeleton className="h-64 rounded-xl" />
      </div>
    );
  }

  return (
    <div className="space-y-4 lg:space-y-6" data-testid="dashboard-page">
      {/* KPI Row */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4" data-testid="dashboard-kpi">
        <KPICard title="Ordini Totali" value={stats?.total_orders || 0} icon={ClipboardList} description={`${stats?.pianificabili || 0} da pianificare`} testId="kpi-total-orders" />
        <KPICard title="In Viaggio" value={stats?.in_viaggio || 0} icon={Truck} description={`${stats?.chiusi || 0} chiusi`} testId="kpi-in-viaggio" />
        <KPICard title="Clienti" value={stats?.total_customers || 0} icon={Users} testId="kpi-customers" />
        <KPICard title="Fatturato" value={`€ ${formatEuro(stats?.total_revenue || 0)}`} icon={FileText} description={`${stats?.fatturati || 0} fatturati`} testId="kpi-revenue" />
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-[1.6fr_1fr] gap-4">
        {/* Monthly Trend Chart */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
              <TrendingUp className="h-4 w-4" /> Andamento Ordini
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-56">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={stats?.monthly_trend || []}>
                  <CartesianGrid strokeDasharray="3 3" stroke="hsl(214 18% 88%)" />
                  <XAxis dataKey="mese" tick={{ fontSize: 11 }} stroke="hsl(215 16% 38%)" />
                  <YAxis tick={{ fontSize: 11 }} stroke="hsl(215 16% 38%)" />
                  <Tooltip
                    contentStyle={{ borderRadius: 8, fontSize: 12, border: '1px solid hsl(214 18% 88%)' }}
                    formatter={(val: number, name: string) => [name === 'totale' ? `€ ${formatEuro(val)}` : val, name === 'totale' ? 'Fatturato' : 'Ordini']}
                  />
                  <Bar dataKey="ordini" fill="#0EA5A6" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Recent Orders */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-base" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Ordini Recenti</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <div className="overflow-x-auto" data-testid="dashboard-recent-orders">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="text-xs py-2">Prog.</TableHead>
                    <TableHead className="text-xs py-2">Cliente</TableHead>
                    <TableHead className="text-xs py-2">Stato</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {orders.slice(0, 8).map((o) => (
                    <TableRow key={o.id} className="hover:bg-muted/60">
                      <TableCell className="text-xs py-2 font-mono">{o.progressivo}</TableCell>
                      <TableCell className="text-xs py-2 max-w-[140px] truncate">{o.cliente?.ragione_sociale}</TableCell>
                      <TableCell className="py-2"><StatusBadge stato={o.stato} /></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
